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
- Тесты:
```bash
go test ./...
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
