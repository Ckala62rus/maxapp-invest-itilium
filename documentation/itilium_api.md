# ITILIUM API в проекте

## Назначение
Этот документ собирает в одном месте:

- текущий контракт backend facade в этом проекте
- legacy-интеграцию из `example`
- реальные URL, параметры и структуры данных, которые нужно учитывать при подключении к тестовому ITILIUM

Документ отвечает на вопросы:

- какой базовый URL и какие переменные окружения используются
- какие параметры и в каком формате отправляются
- какие типы данных ожидаются на входе и выходе
- как текущий backend facade соотносится со старой aiogram-интеграцией

## Конфигурация

Текущий Go backend читает настройки ITILIUM через `internal/config/config.go`:

- `ITILIUM_BASE_URL` - базовый URL ITILIUM для outbound-запросов backend
- `ITILIUM_LOGIN` - логин Basic Auth
- `ITILIUM_PASSWORD` - пароль Basic Auth
- `ITILIUM_TIMEOUT` - timeout внешнего запроса

В legacy-примере `example/telegram_bot_itilium/.env` используются:

- `ITILIUM_URL`
- `ITILIUM_TEST_URL`
- `ITILIUM_LOGIN`
- `ITILIUM_PASSWORD`

Безопасно подтвержденный тестовый URL из aiogram `.env`:

- `https://inv-vsrv-1c.bars.ryazan.ru/itilium-test/hs/TelegramNew/`

Важно:

- тестовый URL можно использовать как reference для локальной настройки
- логин и пароль существуют в `example/telegram_bot_itilium/.env`, но не должны переноситься в коммитируемую документацию как открытые секреты
- в рабочем проекте эти значения должны задаваться только через локальный runtime config / environment variables

## Текущий backend facade

Фронтенд не ходит напрямую в raw ITILIUM endpoints. Он вызывает backend facade вида `/api/v1/...`, а backend уже преобразует эти вызовы в outbound HTTP-запросы через `internal/api/itilium_client.go`.

### Аутентификация и идентификация пользователя в текущем состоянии

Сейчас backend получает acting user id из:

- заголовка `X-User-ID`, или
- query-параметра `userId`

Это реализовано в `internal/middleware/identity.go`.

Это временный переходный механизм. Для корректного MAX Mini App production flow user id должен браться не из доверия к заголовку/параметру клиента, а из валидированного MAX init data / token payload.

## Контракт текущего project API

Все JSON-ответы backend во фронтенд завернуты в общий envelope:

```json
{
  "success": true,
  "message": "optional string",
  "data": {},
  "requestId": "optional string"
}
```

### `GET /api/v1/users/me`

Назначение:

- получить профиль текущего MAX-пользователя, уже приведенный к формату проекта

Источник identity:

- `user_id` из middleware context

Тип `data` в ответе: `UserProfile`

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

Назначение:

- зарегистрировать или привязать пользователя, которого не нашли в ITILIUM

Тип request `data`: `RegistrationRequest`

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

Тип `data` в ответе: `UserProfile`

### `GET /api/v1/tickets`

Назначение:

- получить список собственных заявок текущего пользователя

Источник identity:

- `user_id` из middleware context

Тип `data` в ответе: `TicketSummary[]`

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

Назначение:

- получить список заявок, где текущий пользователь является ответственным

Источник identity:

- `user_id` из middleware context

Тип `data` в ответе: `TicketSummary[]`

### `POST /api/v1/tickets/search`

Назначение:

- найти одну заявку по номеру

Тип request `data`: `SearchTicketRequest`

```json
{
  "number": "string",
  "userId": "string"
}
```

Тип `data` в ответе: `TicketDetail`

### `POST /api/v1/tickets`

Назначение:

- создать новую заявку

Тип request `data`: `CreateTicketRequest`

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

Тип `data` в ответе: `TicketDetail`

### `GET /api/v1/tickets/{number}`

Назначение:

- загрузить полную карточку заявки по номеру

Источник identity:

- `user_id` из middleware context

Тип `data` в ответе: `TicketDetail`

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

Назначение:

- добавить комментарий к заявке

Тип request `data`: `AddCommentRequest`

```json
{
  "userId": "string",
  "message": "string",
  "attachments": ["string"]
}
```

Тип `data` в ответе: `TicketDetail`

### `POST /api/v1/tickets/{number}/status`

Назначение:

- сменить статус заявки

Тип request `data`: `ChangeStatusRequest`

```json
{
  "userId": "string",
  "state": "string",
  "comment": "string",
  "date": "string"
}
```

Тип `data` в ответе: `TicketDetail`

### `GET /api/v1/tickets/{number}/responsibles`

Назначение:

- получить список доступных ответственных по заявке

Источник identity:

- `user_id` из middleware context

Тип `data` в ответе: `ResponsibleOption[]`

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

Назначение:

- назначить нового ответственного

Тип request `data`: `ChangeResponsibleRequest`

```json
{
  "userId": "string",
  "responsibleId": "string"
}
```

Тип `data` в ответе: `TicketDetail`

## Legacy ITILIUM endpoints из `example`

Старый aiogram-проект ходит в raw ITILIUM endpoints напрямую. Эти endpoints важны как source of truth для реальных названий параметров и для маппинга при миграции.

### Поиск пользователя / идентификация

#### `POST find_employee`

Form parameters:
- `telegram: int`
- или другой идентификатор, если `attribute_code` переопределен

Назначение:

- это базовый legacy-endpoint получения информации о текущем пользователе из ITILIUM
- через него бот определяет:
  - найден ли пользователь
  - какие у него есть заявки
  - доступен ли маркетинговый сценарий
  - какие дополнительные признаки и флаги вернул ITILIUM

Используемые поля ответа:
- `UUID: string`
- `servicecalls: string[]`
- `canCreateMarketingRequests: bool`

Что важно про реальный ответ:

- фактический payload пользователя может быть шире, чем три перечисленных поля
- в ответе может приходить массив номеров задач пользователя и дополнительные флаги/признаки доступа
- в legacy-коде точно используются `servicecalls` и `canCreateMarketingRequests`, но это не означает, что других полезных полей нет
- при подключении к реальному test server стоит отдельно зафиксировать полный JSON ответа этого endpoint

Практический вывод для текущего проекта:

- `find_employee` нужно рассматривать как raw endpoint получения текущего пользователя из ITILIUM
- будущий backend MAX Mini App должен уметь маппить не только `UUID`, но и дополнительные флаги, если они потребуются UI или бизнес-логике

Замечание по миграции:

- параметр `telegram` должен быть заменен на валидированный MAX user id

### Обычные Service Call заявки

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

Используемые поля ответа:
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

Используемая структура ответа:

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

### Связанные endpoints, которые еще не перенесены

Эти endpoints есть в `example`, но в текущем Go backend facade пока не реализованы:

- `POST confirm_sc`
- `POST vote_change`
- `GET /listServicesMarketing`
- `GET /listSubdivisionMarketing`
- `POST create_sc_Marketing`

## Маппинг: текущий backend facade -> legacy ITILIUM

| Текущий backend route | Текущий тип запроса | Legacy/example ITILIUM endpoint |
|---|---|---|
| `GET /api/v1/tickets` | context user id | список заявок пользователя по identity |
| `GET /api/v1/tickets/responsible` | context user id | `list_sc_responsible?telegram=...` + ticket detail enrichment |
| `POST /api/v1/tickets/search` | `{ number, userId }` | `find_sc?telegram=...&sc_number=...` |
| `POST /api/v1/tickets` | `CreateTicketRequest` | `create_sc` |
| `POST /api/v1/tickets/{number}/comments` | `AddCommentRequest` | `add_comment?...` |
| `POST /api/v1/tickets/{number}/status` | `ChangeStatusRequest` | `change_state_sc?...` |
| `GET /api/v1/tickets/{number}/responsibles` | context user id | `responsibles_sc?telegram=...&sc_number=...` |
| `POST /api/v1/tickets/{number}/responsible` | `ChangeResponsibleRequest` | `change_responsible_sc?...` |

## Правило миграции identity

Пока MAX init data validation не реализован, проект использует временный `userId`, переданный клиентом.

Целевая схема:

- не доверять `X-User-ID` / raw `userId` от клиента как production-решению
- валидировать MAX init data на backend
- извлекать реальный MAX user id из проверенного/расшифрованного токена
- передавать именно этот MAX user id дальше в profile resolution и ITILIUM requests
- на переходном этапе этот MAX user id заменяет Telegram user id из legacy-бота

## Дополнительные замечания

- текущий Go client использует JSON body на внутреннем API-слое проекта и Basic Auth для outbound ITILIUM
- legacy aiogram-интеграция использует много form/query-based вызовов в raw ITILIUM
- при подключении к реальному test server нужно отдельно сверять naming полей, потому что legacy ITILIUM названия не всегда консистентны, например `shorDescription`
