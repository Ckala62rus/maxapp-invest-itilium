# ITILIUM API Integration

## Purpose
This document consolidates the ITILIUM integration contract used by the current Go backend and the legacy example integration from `example/`.

It is intended to answer:
- which URL and credentials are used
- which request parameters are sent
- which field types are expected
- how the current backend facade maps to legacy ITILIUM endpoints

## Configuration

Current backend integration reads ITILIUM settings from environment variables via `internal/config/config.go`:

- `ITILIUM_BASE_URL` - current Go backend outbound base URL
- `ITILIUM_LOGIN` - Basic Auth login
- `ITILIUM_PASSWORD` - Basic Auth password
- `ITILIUM_TIMEOUT` - outbound timeout

Legacy example integration in `example/telegram_bot_itilium/.env` uses:

- `ITILIUM_URL`
- `ITILIUM_TEST_URL`
- `ITILIUM_LOGIN`
- `ITILIUM_PASSWORD`

Important:
- test credentials exist in `example`, but must not be copied into committed project files
- the current project should receive them through local runtime config only

## Current Backend Facade

The frontend does not call raw ITILIUM endpoints directly. It calls the Go backend facade under `/api/v1/...`, and the backend translates these calls into outbound ITILIUM requests through `internal/api/itilium_client.go`.

### Authentication and user identity in current backend

Today the backend gets the acting user id from:

- request header `X-User-ID`, or
- query parameter `userId`

This behavior is implemented in `internal/middleware/identity.go`.

This is currently a transitional mechanism. For MAX Mini App production flow, the user id must come from validated MAX init data rather than from a client-provided header/query param.

## Current Project API Contract

All frontend-facing JSON responses are wrapped in:

```json
{
  "success": true,
  "message": "optional string",
  "data": {},
  "requestId": "optional string"
}
```

### `GET /api/v1/users/me`

Purpose:
- get current MAX user profile resolved for ITILIUM integration

Identity source:
- middleware context user id

Response `data` type: `UserProfile`

```json
{
  "userId": "string",
  "username": "string",
  "fullName": "string",
  "department": "string",
  "employeeFound": true,
  "registrationRequired": false
}
```

### `POST /api/v1/users/register`

Purpose:
- register/link a user that was not found in ITILIUM

Request `data` type: `RegistrationRequest`

```json
{
  "userId": "string",
  "employeeNumber": "string",
  "fullName": "string",
  "department": "string",
  "phone": "string",
  "comment": "string"
}
```

Response `data` type: `UserProfile`

### `GET /api/v1/tickets`

Purpose:
- list current user's own tickets

Identity source:
- middleware context user id

Response `data` type: `TicketSummary[]`

```json
[
  {
    "number": "string",
    "title": "string",
    "state": "string",
    "deadline": "string",
    "responsibleTeam": "string"
  }
]
```

### `GET /api/v1/tickets/responsible`

Purpose:
- list tickets where current user is the responsible person

Identity source:
- middleware context user id

Response `data` type: `TicketSummary[]`

### `POST /api/v1/tickets/search`

Purpose:
- search one ticket by number

Request `data` type: `SearchTicketRequest`

```json
{
  "number": "string",
  "userId": "string"
}
```

Response `data` type: `TicketDetail`

### `POST /api/v1/tickets`

Purpose:
- create a new ticket

Request `data` type: `CreateTicketRequest`

```json
{
  "userId": "string",
  "requestType": "string",
  "title": "string",
  "description": "string",
  "department": "string",
  "executionDate": "string",
  "attachments": ["string"]
}
```

Response `data` type: `TicketDetail`

### `GET /api/v1/tickets/{number}`

Purpose:
- load full ticket details by number

Identity source:
- middleware context user id

Response `data` type: `TicketDetail`

```json
{
  "number": "string",
  "title": "string",
  "description": "string",
  "state": "string",
  "deadline": "string",
  "responsibleTeam": "string",
  "canChangeResponsible": true,
  "availableStates": ["string"],
  "timeline": [
    {
      "author": "string",
      "message": "string",
      "createdAt": "string"
    }
  ]
}
```

### `POST /api/v1/tickets/{number}/comments`

Purpose:
- add a comment to a ticket

Request `data` type: `AddCommentRequest`

```json
{
  "userId": "string",
  "message": "string",
  "attachments": ["string"]
}
```

Response `data` type: `TicketDetail`

### `POST /api/v1/tickets/{number}/status`

Purpose:
- change ticket status

Request `data` type: `ChangeStatusRequest`

```json
{
  "userId": "string",
  "state": "string",
  "comment": "string",
  "date": "string"
}
```

Response `data` type: `TicketDetail`

### `GET /api/v1/tickets/{number}/responsibles`

Purpose:
- list available assignees for a ticket

Identity source:
- middleware context user id

Response `data` type: `ResponsibleOption[]`

```json
[
  {
    "team": "string",
    "person": "string",
    "role": "string",
    "externalId": "string"
  }
]
```

### `POST /api/v1/tickets/{number}/responsible`

Purpose:
- assign a new responsible person

Request `data` type: `ChangeResponsibleRequest`

```json
{
  "userId": "string",
  "responsibleId": "string"
}
```

Response `data` type: `TicketDetail`

## Legacy ITILIUM Endpoints From `example`

The old aiogram integration calls raw ITILIUM endpoints directly. Those endpoints are still useful as source-of-truth for real parameter naming and migration mapping.

### Identity lookup

#### `POST find_employee`

Form parameters:
- `telegram: int`
- or another identifier field when `attribute_code` is overridden

Used response fields:
- `UUID: string`
- `servicecalls: string[]`
- `canCreateMarketingRequests: bool`

Migration note:
- `telegram` must be replaced with a validated MAX user identifier

### Regular service calls

#### `POST create_sc`

Form parameters:
- `client: string` - employee UUID in ITILIUM
- `shorDescription: string`
- `Description: string`
- `files: string` - JSON-encoded array, optional

#### `POST find_sc?telegram={id}&sc_number={number}`

Query parameters:
- `telegram: int`
- `sc_number: string`

Used response fields:
- `number: string`
- `shortDescription: string`
- `state: string`
- `responsibleTeamTitle: string`
- `deadlineDate: string`
- `description: string`
- `new_state: string[]`
- `change_responsible: bool`

#### `POST add_comment?telegram={id}&source={ticket}&source_type=servicecall&comment_text={text}`

Query parameters:
- `telegram: int`
- `source: string`
- `source_type: string`
- `comment_text: string`
- `files: string` - optional semicolon-separated list

#### `POST change_state_sc?telegram={id}&inc_number={ticket}&new_state={state}`

Query parameters:
- `telegram: int`
- `inc_number: string`
- `new_state: string`

#### `POST change_state_sc?telegram={id}&inc_number={ticket}&new_state={state}&date_inc={date}&comment={comment}`

Query parameters:
- `telegram: int`
- `inc_number: string`
- `new_state: string`
- `date_inc: string`
- `comment: string`

#### `POST responsibles_sc?telegram={id}&sc_number={ticket}`

Query parameters:
- `telegram: int`
- `sc_number: string`

Used response shape:

```json
[
  {
    "responsibleTeamId": "string",
    "responsibleTeamTitle": "string",
    "responsibles": [
      {
        "responsibleEmployeeId": "string",
        "responsibleEmployeeTitle": "string"
      }
    ]
  }
]
```

#### `POST change_responsible_sc?telegram={id}&inc_number={ticket}&responsibleEmployeeId={employeeId}`

Query parameters:
- `telegram: int`
- `inc_number: string`
- `responsibleEmployeeId: string`

### Related but not yet migrated

These endpoints exist in the example but are not yet implemented in the current Go backend facade:

- `POST confirm_sc`
- `POST vote_change`
- `GET /listServicesMarketing`
- `GET /listSubdivisionMarketing`
- `POST create_sc_Marketing`

## Mapping: Current Backend Facade -> Legacy ITILIUM

| Current backend route | Current request type | Legacy/example ITILIUM shape |
|---|---|---|
| `GET /api/v1/tickets` | context user id | list from employee/telegram-linked identity |
| `GET /api/v1/tickets/responsible` | context user id | `list_sc_responsible?telegram=...` + ticket detail enrichment |
| `POST /api/v1/tickets/search` | `{ number, userId }` | `find_sc?telegram=...&sc_number=...` |
| `POST /api/v1/tickets` | `CreateTicketRequest` | `create_sc` |
| `POST /api/v1/tickets/{number}/comments` | `AddCommentRequest` | `add_comment?...` |
| `POST /api/v1/tickets/{number}/status` | `ChangeStatusRequest` | `change_state_sc?...` |
| `GET /api/v1/tickets/{number}/responsibles` | context user id | `responsibles_sc?telegram=...&sc_number=...` |
| `POST /api/v1/tickets/{number}/responsible` | `ChangeResponsibleRequest` | `change_responsible_sc?...` |

## Identity Migration Rule

Until MAX init data validation is implemented, the project still uses a manually supplied `userId`.

Target rule:

- do not trust `X-User-ID` / raw `userId` from the client as the final production solution
- validate MAX init data on the backend
- extract real MAX user id from validated/decrypted token data
- pass that MAX user id further to profile resolution and ITILIUM requests
- for the migration period, this MAX user id replaces the Telegram user id used in the legacy bot

## Notes

- Current Go client uses JSON bodies for the internal project API layer and Basic Auth for outbound ITILIUM requests.
- Legacy aiogram example uses many form/query-based calls to raw ITILIUM endpoints.
- When connecting to the real test server, verify field naming carefully because legacy ITILIUM names are not always consistent, for example `shorDescription`.
