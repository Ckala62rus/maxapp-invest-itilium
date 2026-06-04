# ITILIUM API: понятная памятка

## Что это за документ
Это короткая рабочая памятка:

- к какому ITILIUM мы подключаемся
- какой backend facade уже есть в проекте
- какой legacy API был в aiogram-примере
- что самое важное для первого тестирования на реальном сервере

## Какой сервер использовать

Тестовый ITILIUM из aiogram reference:

- `https://<host>/itilium-test/hs/Max/`

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

- `POST /find_employee`
- в текущем контракте используем поле `id=<MAX user id>`

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

Для MAX Mini App используется реальный MAX user id после backend validation.

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
  "registrationRequired": false,
  "registrationPending": false,
  "statusMessage": "string"
}
```

Что важно:

- при `200` backend возвращает профиль найденного пользователя и frontend открывает главный экран
- при `401` или `404` backend возвращает профиль с `registrationRequired=true`, и frontend открывает экран регистрации
- при `403` backend возвращает профиль с `registrationPending=true`, и frontend показывает статус о том, что заявка еще на рассмотрении
- реальный ответ `find_employee` нужно логировать целиком, чтобы зафиксировать полный контракт 1С

#### `POST /api/v1/users/employee`

Зачем:

- сходить в legacy `find_employee` через наш backend
- посмотреть реальный payload пользователя из ITILIUM
- зафиксировать `UUID`, `servicecalls` и дополнительные флаги

Тело запроса:

```json
{
  "identifier": "100245",
  "attributeCode": "id"
}
```

Что важно:

- если `attributeCode` не передан, backend подставляет `id`
- если `identifier` не передан, backend берет текущий `userId` из context
- этот endpoint сделан как отдельный учебный и исследовательский flow
- он пока не заменяет обычный `GET /api/v1/users/me`

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

Что важно:

- сейчас после отправки формы backend переводит пользователя в состояние `registrationPending=true`
- frontend после успешной отправки показывает, что заявка еще на рассмотрении
- следующий шаг интеграции: заменить локальное pending-состояние реальным запросом в 1С, когда будет подтвержден контракт endpoint регистрации

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
    "creationDate": "string",
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

#### `GET /api/v1/marketing/services`

Зачем:

- получить типы маркетинговых заявок и номер формы шага 4

Ответ:

```json
[
  {
    "code": "design",
    "name": "Дизайн",
    "formNumber": "1",
    "formSchema": {
      "formNumber": "1",
      "title": "Дизайн",
      "fields": [
        {
          "key": "layoutName",
          "label": "Название макета",
          "type": "text",
          "required": true,
          "options": []
        }
      ]
    }
  }
]
```

Важно:

- источник истины по типам и `formNumber` — только 1С
- frontend не хардкодит типы маркетинга

#### `GET /api/v1/marketing/subdivisions`

Зачем:

- получить список подразделений для шага 2

Ответ:

```json
[
  {
    "code": "optional-string",
    "name": "ИВ – Иван Васильевич"
  }
]
```

#### `POST /api/v1/marketing/requests`

Зачем:

- создать маркетинговую заявку по выбранному типу и динамической форме шага 4

Тело запроса:

```json
{
  "serviceCode": "design",
  "formNumber": "1",
  "subdivision": "",
  "executionDate": "",
  "withoutDate": false,
  "formData": {
    "layoutName": "Баннер май",
    "requiredText": "Текст для макета"
  }
}
```

Ответ:

- созданная карточка заявки

Важно:

- `subdivision` и `executionDate` временно отправляются пустыми, потому что поля убраны из UI для live-проверки текущего контракта 1С.
- Если 1С подтвердит обязательность этих полей на практике, нужно вернуть их в UI или договориться о значениях по умолчанию на backend.

#### `GET /api/v1/tickets/{number}`

Зачем:

- загрузить карточку заявки

Ответ:

```json
{
  "number": "string",
  "title": "string",
  "description": "string",
  "creationDate": "string",
  "state": "string",
  "deadline": "string",
  "responsibleEmployee": "string",
  "responsibleTeam": "string",
  "canChangeStatus": true,
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

Без файлов — `application/json` (поле `userId` в теле не доверяем, подставляется из сессии):

```json
{
  "message": "string",
  "attachments": []
}
```

С файлами — `multipart/form-data`: поле `payload` (JSON-строка с `message` и `attachments: []`) и повторяющиеся части `attachments` (как при `POST /api/v1/tickets`). Нужен хотя бы непустой текст **или** одно вложение.

В ITILIUM: `POST /add_comment` с теми же полями формы; при вложениях дополнительно части `files` (как у `create_sc`).

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

#### `POST /api/v1/tickets/{number}/confirm`

Оценка решения (legacy `confirm_sc`). Поле `userId` в теле не доверяем — подставляется из сессии.

Тело запроса:

```json
{
  "mark": 0,
  "comment": "string"
}
```

`mark` — целое от **0** до **5**. Для оценок **0, 1 и 2** комментарий обязателен (и на backend, и в UI).

В ITILIUM: `POST /confirm_sc?telegram={maxUserId}&incident={number}&mark={mark}` и при непустом комментарии `&comment_text=...` (тело запроса пустое, как в example-боте).

## На что это маппится в legacy ITILIUM

- получить пользователя и его заявки: `POST /find_employee` (`id`)
- получить мои заявки по номерам из `servicecalls`: `POST /list_sc` (`id`, `sc_number` через `;`)
- получить список заявок в ответственности: `POST /list_sc_responsible` (`id`, `multipart/form-data`)
- найти карточку заявки: `GET /find_sc` (`id`, `sc_number`)
- добавить комментарий: `POST /add_comment` (`id`, `source`, `source_type=servicecall`, `comment_text`, `multipart/form-data`; при файлах — ещё части `files`)
- сменить статус: `POST /change_state_sc` (`id`, `telegram`, `inc_number`, `new_state`, optional `date_inc`, `comment_text`, `multipart/form-data`)
- получить доступных ответственных: **`POST /responsibles_sc?id=…&sc_number=…`** (параметры в query, тело пустое; на части контуров `GET` даёт 405) **или** fallback `POST` с `telegram`+`sc_number` в query, **или** `multipart/form-data` только с `id` и `sc_number` (без `telegram`, иначе возможна ошибка 1С); ответ — массив **команд** с `responsibles` или плоский список
- сменить ответственного: `POST /change_responsible_sc` (`id`, `telegram`, `inc_number`, `responsibleEmployeeId`, `multipart/form-data`)
- оценить решение: `POST /confirm_sc` (query: `telegram`, `incident`, `mark`, опционально `comment_text`; тело пустое) — проксируется как `POST /api/v1/tickets/{number}/confirm`

## Что еще не перенесено из legacy

Пока еще не реализовано в текущем backend facade:

- `vote_change`

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

## Ответы со стороны 1С (типичное поведение HTTP-сервиса)

Ниже не «официальная спецификация», а наблюдаемые на тестовом контуре паттерны: тело ответа и коды могут отличаться, но их удобно сверять с логами backend (`itilium request details` на уровне DEBUG).

### Общие замечания

- Успешные операции часто возвращают **`200 OK`** с **пустым телом** (`response_body=""`), особенно для `add_comment`, `change_state_sc`, `change_responsible_sc` — факт успеха определяется по коду ответа.
- Ошибки аутентификации к публикации 1С: **`401 Unauthorized`** (иногда с текстом в теле).
- Неверный метод для опубликованного ресурса: **`405 Method Not Allowed`** (например, если вызвать `GET` вместо ожидаемого `POST` с `multipart/form-data`).
- Ошибки бизнес-логики или неверные параметры: **`400 Bad Request`** (тело может быть текстом или JSON — смотреть лог).
- Проблемы на стороне сервера 1С: **`5xx`**.

### `POST /find_employee`

- Успех **`200`**: обычно JSON с полями вроде `UUID`, `servicecalls`, флагов доступа; точный набор полей нужно фиксировать по живому ответу.
- Пользователь не найден / нужна регистрация: часто **`401`** с текстом вида «Пользователь с таким id не найден…» (как строка в теле, не JSON).

### `GET /find_sc`

- Успех **`200`**: JSON-объект карточки заявки, например:

```json
{
  "number": "0000019683",
  "shortDescription": "Различные текущие задачи",
  "description": "…",
  "creationDate": "31.10.2024 11:02:32",
  "deadlineDate": "28.05.2025 14:00:45",
  "responsibleEmployeeTitle": "…",
  "state": "04_В работе",
  "change_status": true,
  "change_responsible": true,
  "new_state": ["08_Закрыто", "06_В ожидании ответа", "…"]
}
```

- Заявка не найдена: обычно **`404`** (или пустое/не JSON тело — смотреть лог).

### `POST /list_sc`, `POST /list_sc_responsible`

- Успех **`200`**: JSON-массив объектов кратких данных **или** массив строк-номеров (для «ответственности»), либо обёртка с ключом `data`/`items` — парсер в backend учитывает несколько вариантов.
- Пустой результат иногда приходит как **`204 No Content`** — тогда тела нет; это нужно учитывать при отладке.

### `POST /add_comment`

- Успех **`200`**: нередко **пустое тело**; дальше backend сам запрашивает актуальную карточку через `find_sc`.
- Долгий ответ (десятки секунд) возможен на стороне 1С — это видно по `duration_ms` в логах, а не как отдельный код.

### `POST /change_state_sc`, `POST /change_responsible_sc`

- Успех **`200`**: часто **пустое** тело; затем обновление карточки через `find_sc`.
- Ошибка бизнес-логики тоже может прийти как **`200`** с JSON-строкой в теле, например `"Не заполнены обязательные параметры"` — backend должен трактовать это как ошибку, а не как успех.
- Для перехода в **«Отложено»** передают **`comment_text`** и **`date_inc`** (`DD.MM.YYYY`, например `30.06.2026`).

### `POST /responsibles_sc` (параметры в строке запроса)

- Типичный вызов: **`POST`** на URL вида `/responsibles_sc?id={userId}&sc_number={номер}` без тела.
- Успех **`200`**: JSON-массив; частый формат — элементы с `responsibleTeamTitle` / `responsibleTeamId` и вложенным массивом `responsibles` (`responsibleEmployeeId`, `responsibleEmployeeTitle`). Допускается плоский список — backend разворачивает оба варианта.

### `POST /create_sc`

- Успех **`200`**: тело может содержать номер созданной заявки или обёртку — разбор в `parseCreateSCResponse`.

### `GET /listServicesMarketing`

- Входной параметр: `id` в query-string, например `/listServicesMarketing?id=40367639`.
- Проверено на тестовом контуре 2026-04-29: `GET` возвращает **`200`**, а `POST` (`application/x-www-form-urlencoded` и `multipart/form-data`) возвращает **`405 Method Not Allowed`**.
- Успешный ответ — JSON-массив услуг с номером формы. Реальный формат:

```json
[
  { "КомпонентаУслуги": "SMM", "НомерФормы": 3 },
  { "КомпонентаУслуги": "Акция", "НомерФормы": 3 },
  { "КомпонентаУслуги": "Дизайн", "НомерФормы": 1 },
  { "КомпонентаУслуги": "Иное", "НомерФормы": 3 },
  { "КомпонентаУслуги": "Мероприятие", "НомерФормы": 2 },
  { "КомпонентаУслуги": "Реклама", "НомерФормы": 3 }
]
```

### `GET /listSubdivisionMarketing`

- Входной параметр: `id` в query-string, например `/listSubdivisionMarketing?id=40367639`.
- Проверено на тестовом контуре 2026-04-29: `GET` возвращает **`200`**, а `POST` (`application/x-www-form-urlencoded` и `multipart/form-data`) возвращает **`405 Method Not Allowed`**.
- Успешный ответ — JSON-массив названий подразделений. Для пользователя `40367639` сейчас 1С вернула пустой массив:

```json
[]
```

### `POST /create_sc_Marketing`

- Отправляется как `multipart/form-data`.
- Общие поля: `id`, `Services`, `Subdivision`, `ExecutionDate` (формат `YYYY-MM-DD`, например `2026-06-04`, когда дата используется), опционально части `files`.
- На 2026-04-29 `Subdivision` и `ExecutionDate` намеренно отправляются пустыми из backend, чтобы проверить live-поведение 1С после упрощения формы в mini app.
- Для услуги `Дизайн` добавляются поля: `LayoutName`, `Size`, `ForWhat`, `RequiredText`, `LayoutFormats`, опционально `LinkToFoto`, `LinkToExamples`.
- Для услуги `Мероприятие` добавляются поля: `ThemeEvent`, `Description`, `Budget`, опционально `LinkToFoto`, `LinkToExamples`.

---

## Логи и трассировка

Чтобы удобно разбирать цепочки в Grafana и docker logs:

- ориентируемся на `request_id` в inbound/outbound логах
- в outbound логах ITILIUM сохраняем: `method`, `url`, `status_code`, `duration_ms`, `request_id`, `user_id`
- для form/multipart запросов сохраняем также отправленные поля (`form_fields`/`urlencoded_body`) и тело ответа (обрезанное)

Типичная успешная цепочка добавления комментария:

1. `POST /api/v1/tickets/{number}/comments` (inbound)
2. `POST /add_comment` (ITILIUM, multipart)
3. `GET /find_sc` (обновление карточки после комментария)
4. `http request completed` с тем же `request_id`
