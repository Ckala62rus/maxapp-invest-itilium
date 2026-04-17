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
   ```

### Backend в Docker: обычный запуск и отладка (Air / Delve / GoLand)

- По умолчанию **`air`** читает **`.air.toml`**: собирает бинарник и **запускает его без Delve** — API на `:3000` доступен сразу, без подключения GoLand.
- Для **удалённой отладки** используйте **`.air.debug.toml`**: Delve в контейнере слушает **`:40000`**, наружу в Windows пробрасывается на **`localhost:40100`** (в `docker-compose.dev.yml`), сборка с `-gcflags="all=-N -l"`. Для `dlv exec` флаг `--continue` должен быть перед путём к бинарнику (`exec --continue ./tmp/maxapp-bot`), иначе headless-сессия ведёт себя как «жду отладчик», и HTTP не поднимается, пока не подключите IDE.
  - Через compose (рекомендуется): в PowerShell перед запуском задайте `$env:AIR_CONFIG='.air.debug.toml'`, затем `docker compose -f docker-compose.dev.yml up -d --force-recreate backend-dev`.
  - В обычный режим вернуться так: `$env:AIR_CONFIG='.air.toml'` и повторить `up -d --force-recreate backend-dev`.
- В GoLand: **Run → Attach to Process** / remote на `localhost:40100` (порт проброшен в `docker-compose.dev.yml`).

### Опционально: переменные только для фронта (Vite)

Создайте файл **`frontend/.env.development.local`** (не коммитьте; шаблон — `frontend/.env.development.example`).

| Переменная | Назначение |
|------------|------------|
| `VITE_PUBLIC_API_BASE_URL` | Базовый URL API. По умолчанию пусто — тогда используются **относительные** пути и срабатывает **proxy** в Vite. Если задать, например, `http://127.0.0.1:3000`, axios пойдёт на бэкенд **напрямую** (CORS для `http://localhost:5173` уже разрешён в `internal/middleware/cors.go`). |
| `VITE_DEBUG_USER_ID` | В режиме `DEV`, если нет bearer-токена, в запросы добавляется заголовок `X-User-ID` для отладки вне MAX (если в конфиге бэкенда включено `auth.allow_debug_identity_headers`). |

После изменения `.env*` перезапустите `npm run dev`.

---

## Полный стек в Docker (как в README)

Если нужен один origin через nginx (`8080`):

```bash
docker compose -f docker-compose.dev.yml up --build -d
```

- UI: `http://localhost:8080`  
- Backend напрямую: `http://localhost:3000`  

В этом режиме Vite крутится **внутри** контейнера `frontend`; для прокси в `vite.config.js` обычно указывают target `http://backend-dev:3000`, а не `127.0.0.1`.

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
| `.air.toml` | корень | Air в контейнере `backend-dev`: сборка и запуск **без** Delve |
| `.air.debug.toml` | корень | то же + Delve на `:40000` для GoLand (см. блок про отладку выше) |
| `config.yml` в корне | локально | не коммитить; если используете свой путь — задайте `CONFIG_PATH` |

---

## Типичные проблемы

- **304 на `http://localhost:5173/api/...`** — API не проксируется на Go; включите proxy в Vite или задайте `VITE_PUBLIC_API_BASE_URL` на реальный адрес бэкенда.  
- **Backend недоступен на 3000** — контейнер `backend-dev` не запущен или порт занят.  
- **CORS при прямом URL на API** — разрешены `http://localhost:5173` и `http://127.0.0.1:5173`; при другом origin добавьте правило в `internal/middleware/cors.go`.
- **В Docker меняете Go-код, а `backend-dev` не пересобирается** — на Docker Desktop (Windows/macOS) события с хоста через bind mount часто не доходят до watcher’а внутри Linux-контейнера. В `.air.toml` включён **`poll = true`** (периодическая проверка файлов), после правок перезапустите контейнер `backend-dev`, чтобы подхватить конфиг. Если бэкенд запускаете **`go run ./cmd/bot` на хосте**, hot reload нет — перезапускайте процесс вручную или используйте `air` локально.
- **API не отвечает, пока не подключите отладку в GoLand** — проверьте `.air.debug.toml`: у `dlv exec` флаг `--continue` должен идти **до** пути к бинарнику (`exec --continue ./tmp/maxapp-bot`), иначе Delve ждёт attach и не запускает HTTP до подключения IDE.
