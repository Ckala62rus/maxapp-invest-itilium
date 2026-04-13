# ITILIUM API: понятная памятка

## Что это за документ
Это короткая рабочая памятка:

- к какому ITILIUM мы подключаемся
- какой backend facade уже есть в проекте
- какой legacy API был в aiogram-примере
- что самое важное для первого тестирования на реальном сервере

## Какой сервер использовать

Тестовый ITILIUM из aiogram reference:

- `https://inv-vsrv-1c.bars.ryazan.ru/itilium-test/hs/TelegramNew/`

Откуда это взято:

- `example/telegram_bot_itilium/.env`

Какие переменные нужны в текущем проекте:

- `ITILIUM_BASE_URL`
- `ITILIUM_LOGIN`
- `ITILIUM_PASSWORD`
- `ITILIUM_TIMEOUT`

Важно:

- URL можно брать из aiogram `.env`
- логин и пароль нужно держать только локально
- секреты не должны попадать в git и в документацию

## Как сейчас устроена интеграция

Сейчас цепочка такая:

1. frontend вызывает наш Go backend по `/api/v1/...`
2. Go backend ходит в ITILIUM через `internal/api/itilium_client.go`

То есть frontend не должен знать raw ITILIUM endpoints.

## Самое важное про пользователя

### Legacy endpoint `find_employee`

Это ключевой endpoint в старой схеме.

Его нужно понимать не как “найти UUID”, а как:

- получить текущего пользователя из ITILIUM
- получить список его заявок
- получить флаги доступа

Формат запроса в legacy:

- `POST find_employee`
- обычно передается `telegram=<id>`

Что точно используется в ответе:

- `UUID`
- `servicecalls`
- `canCreateMarketingRequests`

Что это значит:

- `UUID` - идентификатор сотрудника в ITILIUM
- `servicecalls` - массив номеров/идентификаторов заявок пользователя
- `canCreateMarketingRequests` - можно ли показывать/разрешать маркетинговые заявки

Важно:

- реальный ответ этого endpoint может содержать больше флагов, чем мы уже знаем
- поэтому при первом реальном тесте нужно сохранить полный JSON ответа этого endpoint

Для MAX Mini App этот endpoint в будущем должен использовать:

- не Telegram ID
- а реальный MAX user id после backend validation

## Что уже есть в нашем backend facade

Ниже не raw ITILIUM, а именно то, что уже доступно frontend через наш backend.

### Профиль

#### `GET /api/v1/users/me`

Зачем:

- получить профиль текущего пользователя

Ответ:

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

#### `POST /api/v1/users/register`

Зачем:

- отправить форму регистрации, если пользователя не нашли

Тело запроса:

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

### Заявки

#### `GET /api/v1/tickets`

Зачем:

- получить список моих заявок

Ответ:

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

#### `GET /api/v1/tickets/responsible`

Зачем:

- получить список заявок в моей ответственности

Ответ по структуре такой же, как и у списка моих заявок.

#### `POST /api/v1/tickets/search`

Зачем:

- найти заявку по номеру

Тело запроса:

```json
{
  "number": "string",
  "userId": "string"
}
```

Ответ:

- полная карточка заявки

#### `POST /api/v1/tickets`

Зачем:

- создать новую заявку

Тело запроса:

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

Ответ:

- созданная карточка заявки

#### `GET /api/v1/tickets/{number}`

Зачем:

- загрузить карточку заявки

Ответ:

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

#### `POST /api/v1/tickets/{number}/comments`

Тело запроса:

```json
{
  "userId": "string",
  "message": "string",
  "attachments": ["string"]
}
```

#### `POST /api/v1/tickets/{number}/status`

Тело запроса:

```json
{
  "userId": "string",
  "state": "string",
  "comment": "string",
  "date": "string"
}
```

#### `GET /api/v1/tickets/{number}/responsibles`

Ответ:

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

#### `POST /api/v1/tickets/{number}/responsible`

Тело запроса:

```json
{
  "userId": "string",
  "responsibleId": "string"
}
```

## На что это маппится в legacy ITILIUM

| Что нужно нам | Legacy endpoint |
|---|---|
| получить пользователя и его заявки | `find_employee` |
| создать обычную заявку | `create_sc` |
| найти заявку по номеру | `find_sc` |
| добавить комментарий | `add_comment` |
| сменить статус | `change_state_sc` |
| получить доступных ответственных | `responsibles_sc` |
| сменить ответственного | `change_responsible_sc` |
| получить список заявок в ответственности | `list_sc_responsible` |

## Что еще не перенесено из legacy

Пока еще не реализовано в текущем backend facade:

- `confirm_sc`
- `vote_change`
- `listServicesMarketing`
- `listSubdivisionMarketing`
- `create_sc_Marketing`

## Что важно для первого теста на реальном сервере

Порядок проверки я бы делал такой:

1. сначала проверить `find_employee`
2. посмотреть полный JSON ответа
3. зафиксировать:
   - все флаги
   - массивы
   - поля, которые реально нужны фронту
4. потом уже проверять:
   - мои заявки
   - заявки в ответственности
   - поиск
   - карточку
   - комментарий / статус / ответственного

## Главное ограничение на сегодня

Сейчас `userId` в проекте временный.

Пока не сделана MAX identity validation:

- backend временно доверяет `X-User-ID` / `userId`
- в ITILIUM уходит временный идентификатор

Правильная целевая схема:

- MAX JS library
- init data / token
- backend validate/decrypt
- реальный MAX user id
- этот MAX user id идет дальше в ITILIUM вместо Telegram ID
