# MAX ITILIUM Mini App

Проект переводит Telegram `aiogram`-бота ITILIUM в архитектуру MAX Mini App: `Vue 3` фронт, `Go` backend, Docker dev/prod, Redis, Postgres, Prometheus, Loki и Grafana.

## Что уже есть
- статический адаптивный прототип экранов в `frontend/`
- Go backend scaffold в `cmd/` и `internal/`
- dev/prod `docker-compose`
- mounted config через `deploy/config/*.yml`
- базовые middleware для `request_id`, логирования, метрик и panic recovery
- demo ITILIUM client для безопасной разработки без live API
- первая SQL миграция
- документация по UX, маршрутам, миграциям и общей архитектуре

## Быстрый старт
1. Скопируйте `.env.example` в `.env`.
2. Поднимите dev-стек:
```bash
docker compose -f docker-compose.dev.yml up --build -d
```
3. Откройте:
   - UI: `http://localhost:8080`
   - backend: `http://localhost:3000`
   - Grafana: `http://localhost:3001`
   - Prometheus: `http://localhost:9090`

## Frontend
- Локально:
```bash
cd frontend
npm install
npm run dev
```
- Production build:
```bash
cd frontend
npm run build
```

## Backend
- Локально:
```bash
go run ./cmd/bot
```

## Тесты

Go-тесты вынесены в отдельную директорию `tests/` и сгруппированы по проверяемым слоям:

- `tests/api` — разбор ответов и флагов ITILIUM/1С
- `tests/auth` — MAX initData и backend access token
- `tests/handlers` — HTTP routes/middleware
- `tests/services` — бизнес-логика сервисов

### Запуск всех Go-тестов

Эта команда проходит по всем пакетам проекта: `cmd`, `internal` и `tests`.
Пакеты без тестов будут показаны как `[no test files]`, это нормально.

```bash
go test ./...
```

Если нужно отключить кеш тестов Go и гарантированно выполнить всё заново:

```bash
go test ./... -count=1
```

Подробный вывод с именами тестов и подтестов:

```bash
go test ./... -v
```

### Запуск только директории `tests`

Эта команда запускает только вынесенные тесты, не заходя в пакеты приложения без тестовых файлов:

```bash
go test ./tests/...
```

### Запуск конкретной папки

```bash
go test ./tests/api
go test ./tests/services
go test ./tests/auth
go test ./tests/handlers
```

С подробным выводом и без кеша:

```bash
go test ./tests/services -v -count=1
```

### Запуск конкретного теста

Флаг `-run` принимает регулярное выражение по имени теста:

```bash
go test ./tests/auth -run TestManagerValidateInitData -v
```

Можно запускать один тест из группы сервисов:

```bash
go test ./tests/services -run TestProfileServiceGetProfileFallback -v
```

### Запуск конкретного подтеста

В table-driven тестах подтесты запускаются через полный путь `Тест/Подтест`.
Например, проверить только кейс `russian_truthy_string`:

```bash
go test ./tests/api -run 'TestMarketingPermissionFromFindEmployeePayload_synonyms/russian_truthy_string' -v
```

В PowerShell одинарные кавычки тоже подходят:

```powershell
go test ./tests/api -run 'TestMarketingPermissionFromFindEmployeePayload_synonyms/russian_truthy_string' -v
```

Frontend-сборка для проверки Vue-кода:

```bash
cd frontend
npm run build
```

## Миграции
Применить миграции:
```bash
docker compose -f docker-compose.dev.yml --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" up
```

Подробности см. в `documentation/migrations_and_mocks.md`.

## Документация
- `documentation/local_development.md` — локальная разработка (env, Vite, proxy, `go run`)
- `documentation/aiogram_feature_map.md`
- `documentation/ui_flows.md`
- `documentation/system_overview.md`
- `documentation/routes.md`
- `documentation/migrations_and_mocks.md`

## Наблюдаемость
- Логи идут в stdout и могут собираться Loki driver'ом Docker.
- Метрики отдаются на `/metrics`.
- Grafana и Prometheus уже добавлены в compose.

## Следующий шаг
Следующий логичный этап после согласования прототипа с заказчиком: перевод экранов в полноценную Vue 3 модульную структуру и подключение live backend сценариев к реальному контракту ITILIUM.
