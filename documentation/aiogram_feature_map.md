# AIogram Bot Feature Map

## Purpose
This document decomposes the legacy Telegram `aiogram` bot into business scenarios that can be migrated into the MAX Mini App without copying the Python structure as-is.

## Source References
- `example/telegram_bot_itilium/src/handlers/new_user_handler.py`
- `example/telegram_bot_itilium/src/api/itilium_api.py`
- `example/telegram_bot_itilium/src/services/user_private_service.py`
- `example/telegram_bot_itilium/src/dialogs/bot_menu/`
- `example/maxapp_bot_mini_app/documentation/itilium_api.md`

## High-Level Functional Areas

### 1. Entry and Navigation
- `/start` initializes the conversation and sends the base menu.
- `/menu` or the `Меню` button opens the main command hub.
- The menu is the root of all supported user scenarios.

### 2. User Identification in ITILIUM
- Almost every business action starts with `find_employee`.
- The legacy bot uses Telegram user id as the main identifier.
- If the employee is not found, the bot blocks the scenario and shows an error.
- In the MAX Mini App this step becomes a separate use case:
  - validate MAX init data
  - resolve user in ITILIUM
  - if not found, open registration flow

### 3. Create Regular Ticket
- The user opens the create ticket scenario from the main menu.
- The bot asks for a description.
- The user may attach files.
- The bot calls `create_sc`.
- The request payload is based on:
  - `client`
  - `shorDescription`
  - `Description`
  - `files`

### 4. Create Marketing Ticket
- If ITILIUM says `canCreateMarketingRequests=true`, the user can choose between:
  - regular IT ticket
  - marketing request
- The marketing flow contains multiple steps:
  - choose service
  - choose subdivision
  - choose execution date
  - fill service-specific fields
  - optionally attach files
  - submit the request
- Marketing requests use different endpoints and have request shape variations by service type.

### 5. My Tickets
- The bot loads the employee profile from ITILIUM.
- It reads the `servicecalls` list from the employee payload.
- Then it loads each ticket in detail through `find_sc`.
- In the new application this should become:
  - paginated backend endpoint
  - optionally cached ticket projections in Redis
  - separate summary DTO for list pages

### 6. Tickets in My Responsibility
- The bot calls `list_sc_responsible`.
- Then it resolves each returned ticket id using `find_sc`.
- This is a dedicated list page in the MAX Mini App with filters, pagination and card layout.

### 7. Search Ticket by Number
- The user enters a ticket number manually.
- The bot calls `find_sc`.
- The result page shows full ticket data and actions available for this user.

### 8. Ticket Details and Actions
- The details view is the operational center of the legacy bot.
- After opening a ticket, the user may:
  - view main fields
  - add a comment
  - change state
  - vote accept or reject
  - rate a completed ticket
  - change responsible person when ITILIUM allows it

### 9. Commenting
- The user enters text and may add files.
- The bot sends the comment via `add_comment`.
- The MAX Mini App should split this into:
  - comment compose form
  - file uploader
  - submit action
  - response panel

### 10. Status Change
- The bot loads available next states from `find_sc`.
- If a selected state requires additional data, it asks for comment and date.
- Otherwise it calls `change_sc_state` directly.
- The new app should model this as:
  - available actions from ticket details
  - modal or inline form for state transition
  - dynamic fields based on selected state

### 11. Voting and Confirmation
- The legacy bot supports:
  - approval voting through `vote_change`
  - ticket confirmation and rating through `confirm_sc`
- These are separate action widgets in the new UI because their business intent is different from regular status change.

### 12. Change Responsible
- The bot fetches responsible teams and employees.
- It shows paginated options.
- The user selects a new responsible person.
- The bot calls `change_responsible`.
- In the MAX Mini App this needs a dedicated selector dialog/page with backend pagination.

## Migration Map

| Legacy Scenario | New Backend Module | New Frontend Screen |
| --- | --- | --- |
| `/start`, `/menu` | `auth`, `navigation` | home dashboard |
| `find_employee` check | `auth`, `registration` | bootstrap loader, registration page |
| create regular issue | `tickets` | create ticket page |
| create marketing issue | `marketing_tickets` | marketing form wizard |
| my tickets | `my_tickets` | my tickets list |
| responsible tickets | `responsible_tickets` | responsible tickets list |
| search by number | `ticket_search` | search page |
| show ticket info | `ticket_details` | ticket details page |
| add comment | `ticket_comments` | comment form block |
| change status | `ticket_workflow` | ticket action panel |
| vote accept/reject | `ticket_workflow` | action banner / modal |
| confirm with rating | `ticket_workflow` | rating modal |
| change responsible | `ticket_workflow` | responsible selector |

## Proposed Route Skeleton
- `POST /api/v1/auth/max/validate`
- `GET /api/v1/users/me`
- `POST /api/v1/users/register`
- `GET /api/v1/tickets`
- `GET /api/v1/tickets/responsible`
- `POST /api/v1/tickets/search`
- `GET /api/v1/tickets/{number}`
- `POST /api/v1/tickets`
- `POST /api/v1/tickets/{number}/comments`
- `POST /api/v1/tickets/{number}/status`
- `POST /api/v1/tickets/{number}/vote`
- `POST /api/v1/tickets/{number}/confirm`
- `GET /api/v1/tickets/{number}/responsibles`
- `POST /api/v1/tickets/{number}/responsible`
- `GET /api/v1/marketing/services`
- `GET /api/v1/marketing/subdivisions`
- `POST /api/v1/marketing/requests`

## ITILIUM Endpoints To Preserve
- `find_employee`
- `create_sc`
- `find_sc`
- `add_comment`
- `confirm_sc`
- `vote_change`
- `list_sc_responsible`
- `change_sc_state`
- `change_sc_state_with_comment`
- `get_responsibles`
- `change_responsible`
- `listServicesMarketing`
- `listSubdivisionMarketing`
- `create_sc_Marketing`

## Design Notes
- The MAX Mini App should not copy Telegram FSM logic one-to-one.
- State must move from chat-driven steps to explicit page state and form state.
- ITILIUM integrations should be wrapped behind service interfaces and client interfaces.
- Expensive list/detail lookups should be measured and optionally cached in Redis.
- Every request to ITILIUM must be logged with request id, user id, endpoint, latency and response status.
