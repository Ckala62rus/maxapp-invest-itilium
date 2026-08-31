# Локальная разработка (Windows / хост)

Ниже — рабочий вариант «фронт на хосте + API и инфраструктура в Docker», который обычно используется для ежедневной разработки. Также кратко описаны альтернативы.

## Рекомендуемый сценарий: Vite на хосте, backend в Docker

1. **Корень репозитория — `.env` для Compose**  
   Скопируйте `.env.example` → `.env` и при необходимости заполните переменные (ITILIUM, `MAX_BOT_TOKEN`, секреты auth и т.д.). Этот файл читает `docker compose` при `docker-compose.dev.yml`.

2. **Поднимите инфраструктуру и backend** (Postgres, Redis, API на порту **3000**):
   ```bash
   docker compose -f docker-compose.dev.yml up -d postgres redis backend-dev
   ```
   Сервис `frontend` в compose для этого сценария **не обязателен** — UI запускается через `npm run dev` на машине.

3. **Фронтенд на хосте**:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   Откройте `http://localhost:5173`. Запросы к `/api`, `/healthz`, `/readyz`, `/metrics` проксируются на `http://127.0.0.1:3000` (см. `frontend/vite.config.js`). Без этого прокси браузер стучится в сам Vite (`5173`), а не в Go.

4. **Проверка API**:
   ```bash
   curl http://127.0.0.1:3000/healthz
   curl -i http://127.0.0.1:3000/api/v1/marketing/services
   ```
   Второй вызов **без** `Authorization` / `X-User-ID` должен вернуть **`401 Unauthorized`**, но **не** `404`: если получаете **`404 Not Found`**, значит на `:3000` отвечает **старый процесс без новых маршрутов**. Перезапустите backend (например `docker compose -f docker-compose.dev.yml up -d --build backend-dev`) или пересоберите `go run`/`air`.

### Backend в Docker: единый dev-сервис с Delve (Air / GoLand)

- `backend-dev` в `docker-compose.dev.yml` всегда запускается через **`.air.debug.toml`**: сборка с `-gcflags="all=-N -l"` и `dlv` в headless-режиме.
- Delve слушает **`:40000`** в контейнере и проброшен на хост как **`localhost:40100`**.
- Для `dlv exec` флаг `--continue` должен быть перед путём к бинарнику (`exec --continue ./tmp/maxapp-bot`), иначе headless-сессия ждёт attach и HTTP не поднимается.
- В GoLand используйте **Run → Attach to Process** / remote на `localhost:40100`.

### Опционально: переменные только для фронта (Vite)

Создайте файл **`frontend/.env.development.local`** (не коммитьте; шаблон — `frontend/.env.development.example`).

| Переменная | Назначение |
|------------|------------|
| `VITE_PUBLIC_API_BASE_URL` | Базовый URL API. По умолчанию пусто — тогда используются **относительные** пути и срабатывает **proxy** в Vite. Если задать, например, `http://127.0.0.1:3000`, axios пойдёт на бэкенд **напрямую** (CORS для `http://localhost:5173` уже разрешён в `internal/middleware/cors.go`). |
| `VITE_DEBUG_USER_ID` | В режиме `DEV`, если нет bearer-токена, в запросы добавляется заголовок `X-User-ID` для отладки вне MAX (если в конфиге бэкенда включено `auth.allow_debug_identity_headers`). |

После изменения `.env*` перезапустите `npm run dev`.

---

## Полный стек в Docker (как в README)

```bash
docker compose -f docker-compose.dev.yml up --build -d
```

| URL | Назначение |
|-----|------------|
| **`http://localhost:5173`** | **Основной UI для разработки на Windows.** Vite напрямую, HMR стабилен, `/api` проксируется на `backend-dev`. |
| `http://localhost:8080` | Один origin через nginx (как в production). На Windows возможны **самопроизвольные перезагрузки** из‑за WebSocket HMR через прокси — для ежедневной работы используйте `:5173`. |
| `http://localhost:3000` | Backend API напрямую |

Vite в контейнере `frontend` проксирует `/api` на `http://backend-dev:3000` (`VITE_API_PROXY_TARGET`).

Если нужен стабильный `:8080` без перезагрузок, задайте в сервисе `frontend` переменную `VITE_DISABLE_HMR=true` и обновляйте страницу вручную (Ctrl+F5).

---

## Backend через `go run` на хосте

По умолчанию `cmd/bot` грузит **`deploy/config/app.dev.yml`**, если не задан `CONFIG_PATH`.

В YAML для Docker указаны хосты `postgres` и `redis`. На хосте без Docker их нужно переопределить **переменными окружения** (полный список см. `internal/config/config.go`, функция `applyEnvOverrides`), например:

| Переменная | Пример для хоста |
|------------|------------------|
| `POSTGRES_HOST` | `127.0.0.1` |
| `POSTGRES_PORT` | `5432` |
| `REDIS_ADDRESS` | `127.0.0.1:6379` |
| `HTTP_HOST` / `HTTP_PORT` | при необходимости сменить bind |
| `CONFIG_PATH` | путь к своему YAML, если не `deploy/config/app.dev.yml` |

Секреты и токены так же можно задавать через env (см. `.env.example`: `MAX_BOT_TOKEN`, `AUTH_*`, `ITILIUM_*` и др.) — они перекрывают значения из файла.

Убедитесь, что Postgres и Redis доступны (локально или проброшенные из compose).

---

## Краткая шпаргалка по файлам

| Файл | Где | Зачем |
|------|-----|--------|
| `.env` | корень репозитория | Docker Compose, общие секреты для контейнеров |
| `frontend/.env.development.local` | `frontend/` | только Vite (`VITE_*`), локальные оверрайды |
| `deploy/config/app.dev.yml` | репозиторий | базовый dev-конфиг Go (путь по умолчанию для `go run`) |
| `.air.toml` | корень | альтернативный конфиг Air без Delve (ручной запуск вне compose при необходимости) |
| `.air.debug.toml` | корень | текущий основной конфиг для `backend-dev` в compose: Air + Delve на `:40000` |
| `config.yml` в корне | локально | не коммитить; если используете свой путь — задайте `CONFIG_PATH` |

---

## Типичные проблемы

- **304 на `http://localhost:5173/api/...`** — API не проксируется на Go; включите proxy в Vite или задайте `VITE_PUBLIC_API_BASE_URL` на реальный адрес бэкенда.  
- **Backend недоступен на 3000** — контейнер `backend-dev` не запущен или порт занят.  
- **CORS при прямом URL на API** — разрешены `http://localhost:5173` и `http://127.0.0.1:5173`; при другом origin добавьте правило в `internal/middleware/cors.go`.
- **В Docker меняете Go-код, а `backend-dev` не пересобирается** — на Docker Desktop (Windows/macOS) события с хоста через bind mount часто не доходят до watcher’а внутри Linux-контейнера. В `.air.toml` включён **`poll = true`** (периодическая проверка файлов), после правок перезапустите контейнер `backend-dev`, чтобы подхватить конфиг. Если бэкенд запускаете **`go run ./cmd/bot` на хосте**, hot reload нет — перезапускайте процесс вручную или используйте `air` локально.
- **API не отвечает, пока не подключите отладку в GoLand** — проверьте `.air.debug.toml`: у `dlv exec` флаг `--continue` должен идти **до** пути к бинарнику (`exec --continue ./tmp/maxapp-bot`), иначе Delve ждёт attach и не запускает HTTP до подключения IDE.

---

## Самопроизвольная перезагрузка страницы (как F5)

### Симптомы

- Страница сама обновляется, даже если вы ничего не нажимаете.
- Снова появляется «Проверяем MAX-сессию…» и сбрасывается экран (например, с «Мои заявки» на главную).
- На **`http://localhost:5173`** такого **нет**, а на **`http://localhost:8080`** — **есть**.

### Как проверить, что это именно перезагрузка документа

1. Откройте DevTools → вкладка **Network**.
2. Включите **Preserve log**.
3. Подождите, пока страница «сама» обновится.
4. Если в списке снова появился запрос **`index.html`** (тип **document**) — это полная перезагрузка, а не обычное переключение экранов в Vue.

### Причина

В dev-режиме Vite держит WebSocket для **HMR** (hot reload). На Windows при доступе через **nginx (`:8080`)** WebSocket часто обрывается; клиент `@vite/client` вызывает `location.reload()`. На **`http://localhost:5173`** браузер подключается к Vite напрямую — HMR стабилен.

Дополнительно: если в логах `maxapp-itilium-frontend-dev` есть  
`[vite] http proxy error: /api/... Error: connect ECONNREFUSED 127.0.0.1:3000`,  
значит API из контейнера Vite идёт не на backend. Должно быть `VITE_API_PROXY_TARGET=http://backend-dev:3000` в `docker-compose.dev.yml` (уже настроено в репозитории).

### Решение (выберите одно)

| Вариант | Действие | Когда использовать |
|---------|----------|-------------------|
| **A (рекомендуется)** | Работайте на **`http://localhost:5173`** | Ежедневная разработка на Windows |
| **B** | Отключить HMR для `:8080` | Нужен один origin через nginx |

**Вариант B** — в `docker-compose.dev.yml`, сервис `frontend`, добавьте:

```yaml
environment:
  VITE_API_PROXY_TARGET: http://backend-dev:3000
  VITE_DISABLE_HMR: "true"
```

Пересоздайте контейнер:

```bash
docker compose -f docker-compose.dev.yml up -d --force-recreate frontend
```

После правок фронта обновляйте страницу вручную (**Ctrl+F5**). Hot reload не будет работать и на `:5173`, пока контейнер один и тот же — для HMR используйте вариант A без `VITE_DISABLE_HMR`.

---

## Проверка deep link «Открыть заявку» (кнопка в боте)

Кнопка `open_app` в личке MAX передаёт в mini app строку `start_param`, например `ticket_0000024311`. Frontend при старте открывает **карточку заявки** с этим номером.

### Шаг 0. Подготовка

1. Поднимите стек: `docker compose -f docker-compose.dev.yml up -d`
2. В **`frontend/.env.local`** или **`frontend/.env.development.local`** задайте ваш MAX id для отладки без WebView:

   ```env
   VITE_DEBUG_USER_ID=40367639
   ```

   (подставьте свой id из логов backend после входа в MAX: `user_id=...`)

3. Перезапустите frontend после изменения `.env*`:

   ```bash
   docker compose -f docker-compose.dev.yml restart frontend
   ```

4. В браузере сделайте **Ctrl+F5** — иначе может грузиться старый JS без поддержки `?startapp=`.

### Способ 1. Браузер на `localhost:5173` (быстро, без MAX)

Эмулирует то, что MAX передаёт через `start_param` после нажатия кнопки:

```
http://localhost:5173/?startapp=ticket_0000024311
```

Подставьте реальный номер заявки из ITILIUM (как в `find_sc` / списке «Мои заявки»).

**Ожидаемое поведение:**

1. Кратко «Проверяем MAX-сессию…»
2. Открывается экран **«Карточка заявки»**, не главная
3. В Console (F12) — `[nav] startup deep link ticket { ticketNumber: "0000024311" }`
4. В Network — `GET /api/v1/tickets/0000024311` со статусом **200**
5. На карточке видны тема и статус заявки

**Если остались на главной:**

| Проверка | Что сделать |
|----------|-------------|
| Старый JS в кэше | **Ctrl+F5** или DevTools → Network → Disable cache |
| Нет debug user id | Задайте `VITE_DEBUG_USER_ID` в `frontend/.env.local`, перезапустите `frontend` |
| Просроченный token | DevTools → Application → Local Storage → удалите `access_token`, обновите страницу |
| Неверный формат | Только `ticket_` + номер, без двоеточия (`ticket:…` MAX не принимает) |
| В Console `[nav] deep link skipped` | Смотрите поля `bootstrapOk`, `employeeFound` — auth не прошёл |

**Если карточка пустая / ошибка** — заявка недоступна этому пользователю в ITILIUM или неверный номер; проверьте тот же номер через «Поиск» в меню приложения.

### Способ 2. Кнопка-ссылка в MAX → браузер на localhost (только ПК)

Кнопка **`open_app`** открывает **mini app по URL из настроек бота**, а не произвольный адрес.  
Чтобы по нажатию открыть **`http://localhost:5173/?startapp=...` в обычном браузере**, используйте кнопку типа **`link`**:

```bash
cd tools/max-notify
python notify.py "Тест deep link в браузере" ^
  --link-url "http://localhost:5173/?startapp=ticket_0000024311" ^
  --button-text "Открыть заявку 0000024311"
```

(В PowerShell `^` — перенос строки; в одну строку можно без него.)

**Ограничения:**

- Работает, если MAX **на том же компьютере**, где крутится Vite (`localhost` = ваш ПК).
- На **телефоне** `localhost` указывает на сам телефон — нужен HTTPS-туннель, например  
  `--link-url "https://maxapp-dev.ru.tuna.am/?startapp=ticket_0000024311"`.
- Кнопка `link` открывает **браузер**, не WebView mini app — для отладки UI это нормально.

### Способ 3. Реальная кнопка open_app в MAX (mini app end-to-end)

MAX на телефоне **не откроет** `http://localhost:5173` — нужен **публичный HTTPS URL** mini app в настройках бота.

1. **Туннель на ваш Vite** (пример с [tuna.am](https://tuna.am)):

   ```bash
   tuna http 5173 --subdomain=maxapp-dev
   ```

   Получите URL вида `https://maxapp-dev.ru.tuna.am` (домен `.ru.tuna.am` уже разрешён в `vite.config.js`).

2. **В настройках MAX-бота** укажите URL mini app = этот HTTPS-адрес (со слэшем в конце по документации MAX).

3. **Отправьте тестовое уведомление** с кнопкой (из корня репозитория):

   ```bash
   cd tools/max-notify
   copy .env.example .env
   ```

   Заполните `.env`: `MAX_BOT_TOKEN`, `MAX_NOTIFY_USER_ID` (ваш MAX id). Проверка:

   ```bash
   python notify.py --check
   ```

   Отправка с кнопкой на заявку `0000024311`:

   ```bash
   python notify.py --template assigned --ticket 0000024311
   ```

   Или свой текст:

   ```bash
   python notify.py "Тест deep link" --ticket 0000024311
   ```

4. В приложении **MAX** откройте личку с ботом → нажмите **«Открыть заявку»**.

5. **Ожидаемое поведение:** открывается mini app по HTTPS-туннелю → сразу карточка `0000024311`.

**Если открывается главная без заявки:**

- В DevTools (remote debug WebView) или в блоке «MAX bridge debug» на главной проверьте, что `initDataUnsafe.start_param` = `ticket_0000024311`
- Убедитесь, что задеployен frontend с поддержкой `start_param` (файлы `maxBridge.js`, `App.vue`)
- URL бота должен указывать на **тот же** стенд, куда вы деплоите код (не production, если тестируете локально)

Подробнее про API кнопки и Python-примеры: [`tools/max-notify/README.md`](../tools/max-notify/README.md).
