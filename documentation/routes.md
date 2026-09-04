# Route Documentation

## Middleware Chain
Every request passes through the same middleware chain defined in `internal/handlers/routes.go`.

1. `RequestID`
   - Creates or reuses `X-Request-ID`
   - Stores it in request context
2. `Identity`
   - Extracts user id from `X-User-ID` or `userId`
   - Stores it in request context
3. `Recover`
   - Converts panics into structured `500` responses
4. `Metrics`
   - Collects Prometheus request counters and latency
5. `Logging`
   - Logs request and response metadata

## Infrastructure Routes

### `GET /healthz`
- Handler: `Handler.Health`
- Purpose: liveness probe
- Response: simple `ok`

### `GET /readyz`
- Handler: `Handler.Ready`
- Purpose: readiness probe
- Response: simple `ready`

### `GET /metrics`
- Handler: Prometheus `promhttp.Handler`
- Purpose: metrics scraping

## API Routes

### `GET /api/v1/users/me`
- Handler: `Handler.GetProfile`
- Service: `ProfileService.GetProfile`
- Flow:
  - reads `user_id` from middleware context
  - loads user profile from repository
  - if no profile exists, returns fallback profile that requires registration

### `POST /api/v1/users/register`
- Handler: `Handler.RegisterUser`
- Service: `ProfileService.Register`
- Flow:
  - decodes registration form
  - validates required fields
  - stores registration through repository
  - returns updated profile snapshot

### `GET /api/v1/tickets`
- Handler: `Handler.ListMyTickets`
- Service: `TicketService.ListMyTickets`
- Flow:
  - reads `user_id`
  - calls the ITILIUM client
  - returns a list for the `myTickets` screen

### `GET /api/v1/tickets/responsible`
- Handler: `Handler.ListResponsibleTickets`
- Service: `TicketService.ListResponsibleTickets`
- Flow:
  - reads `user_id`
  - calls the ITILIUM client
  - returns a list for the `responsible` screen

### `POST /api/v1/tickets/search`
- Handler: `Handler.SearchTicket`
- Service: `TicketService.SearchTicket`
- Flow:
  - decodes ticket number payload
  - calls the ITILIUM client search endpoint
  - returns the full ticket card

### `POST /api/v1/tickets`
- Handler: `Handler.CreateTicket`
- Service: `TicketService.CreateTicket`
- Flow:
  - decodes create form
  - validates minimal required fields
  - calls ITILIUM `POST /create_sc` (IT) or `POST /create_sc_Dax` (тип «Заявка в DAX») — одинаковые поля multipart
  - returns the created ticket

### `GET /api/v1/tickets/{number}`
- Handler: `Handler.GetTicket`
- Service: `TicketService.GetTicket`
- Flow:
  - checks Redis cache first
  - if cache miss, calls the ITILIUM client
  - returns the full ticket card

### `GET /api/v1/tickets/{number}/comments`
- Handler: `Handler.ListComments`
- Service: `TicketService.ListComments`
- ITILIUM: `GET /list_comment` with query `id`, `sc_number`
- Flow:
  - reads ticket number and identity user id
  - calls ITILIUM `list_comment`
  - maps `sender`/`comment`/`date_sending` → `author`/`message`/`createdAt`
  - returns comment list for the ticket card

### `POST /api/v1/tickets/{number}/comments`
- Handler: `Handler.AddComment`
- Service: `TicketService.AddComment`
- Flow:
  - decodes comment payload
  - validates comment body
  - calls ITILIUM to add the comment
  - returns updated ticket detail

### `POST /api/v1/tickets/{number}/status`
- Handler: `Handler.ChangeStatus`
- Service: `TicketService.ChangeStatus`
- Flow:
  - decodes status transition payload
  - validates target state
  - calls ITILIUM to change workflow state
  - returns updated ticket detail

### `GET /api/v1/tickets/{number}/responsibles`
- Handler: `Handler.ListResponsibleOptions`
- Service: `TicketService.ListResponsibleOptions`
- Flow:
  - reads ticket number
  - calls ITILIUM for available responsible people
  - returns the selector list used by the ticket card

### `POST /api/v1/tickets/{number}/responsible`
- ITILIUM: `change_responsible_sc` with `id`, `inc_number`, `responsibleEmployeeId` and/or `responsibleTeamId`
- Handler: `Handler.ChangeResponsible`
- Service: `TicketService.ChangeResponsible`
- Flow:
  - decodes selected responsible person id
  - validates payload
  - calls ITILIUM to change assignee
  - returns updated ticket detail

### `POST /api/v1/tickets/{number}/confirm`
- Handler: `Handler.ConfirmTicket`
- Service: `TicketService.ConfirmTicket`
- Flow:
  - decodes rating `mark` (0–5) and optional `comment` (required in service for marks 0–2)
  - calls ITILIUM `confirm_sc` (legacy POST with query `telegram`, `incident`, `mark`, optional `comment_text`)
  - returns refreshed ticket detail from `find_sc`

### `GET /api/v1/marketing/services`
- Handler: `Handler.ListMarketingServices`
- Service: `TicketService.ListMarketingServices`
- Middleware: `RequireIdentity` (Bearer или debug `X-User-ID`)
- Flow: читает маркетинговые типы и `formNumber`/схему из ITILIUM (`GET /listServicesMarketing?id=...`; 1С возвращает `КомпонентаУслуги` + `НомерФормы`).

### `GET /api/v1/marketing/subdivisions`
- Handler: `Handler.ListMarketingSubdivisions`
- Service: `TicketService.ListMarketingSubdivisions`
- Middleware: `RequireIdentity`
- Flow: подразделения для шага 2 (`GET /listSubdivisionMarketing?id=...`; успешный ответ — массив названий подразделений, сейчас на тестовом контуре для `40367639` приходит `[]`).

### `POST /api/v1/marketing/requests`
- Handler: `Handler.CreateMarketingRequest`
- Service: `TicketService.CreateMarketingRequest`
- Middleware: `RequireIdentity`
- Flow: создаёт маркетинговую заявку через `POST /create_sc_Marketing` (сначала query: `id`, `Services`, `Subdivision`, `ExecutionDate`, `FormNumber`, `Description`; с файлами — multipart + `files`).
