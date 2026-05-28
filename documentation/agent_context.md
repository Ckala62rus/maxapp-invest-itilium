# Working Context

## Current State

- Dev backend runs through a single `backend-dev` service in `docker-compose.dev.yml`; it starts Air with `.air.debug.toml` (Delve on container `:40000`, published to host `localhost:40100`), while `.air.toml` remains an optional non-Delve fallback for manual runs; Air polling is enabled for Docker Desktop bind mounts.
- Docker containers are currently started by the user and ready for the next external integration checks.
- The ITILIUM dev target is built from `ITILIUM_HOST` in `.env` through `docker-compose.dev.yml`.
- TLS verification is temporarily disabled with `itilium.insecure_skip_verify: true` because the current test host is addressed by IP.
- The legacy user lookup contract was updated to `POST /find_employee` under `/itilium-test/hs/Max/` with form field `id`.
- Live verification on 2026-04-15 confirmed `GET /api/v1/users/me` reaches ITILIUM and currently maps the upstream `401` for user `100245` into `registrationRequired=true`.
- A new Telegram bot was created by the user and is currently waiting for moderation before end-to-end mini app checks can start.
- The app now has a first MAX auth slice: frontend loads `max-web-app.js`, reads `window.WebApp.initData`, exchanges it through `POST /api/v1/auth/max/validate`, and then uses a backend bearer token for protected API calls.
- Live verification on 2026-04-15 confirmed MAX auth validation works with real `initData`; backend logs show `user_id=40367639` and then `find_employee` in ITILIUM returns `401` with `Пользователь с таким id не найден.Необходима регистрация`.
- Ticket legacy contracts were aligned with 1C endpoints under `/hs/Max/`: `list_sc`, `list_sc_responsible`, `find_sc`, `add_comment`, `change_state_sc`, `responsibles_sc`, `change_responsible_sc`.
- Outbound ITILIUM logs now include `request_id` and `user_id`, which allows end-to-end correlation between inbound API calls and downstream 1C calls in Grafana/Loki.

## Implemented Flows

- `POST /api/v1/users/employee` calls legacy `find_employee` and returns normalized lookup data plus `raw`.
- `GET /api/v1/users/me` now uses `find_employee` when there is no confirmed stored registration profile.
- `GET /api/v1/users/me` interprets 1C status codes as onboarding states: `200 -> found`, `401/404 -> registration required`, `403 -> registration pending`.
- `POST /api/v1/users/register` now calls the real ITILIUM endpoint `POST /registration` under `/itilium-test/hs/Max/` with form fields `id`, `FIO`, `Organization`, `Subdivision`, `NamePosition`, and then stores local `registrationPending=true` after a successful upstream response.
- The frontend profile screen shows `MAX ID`, ITILIUM fields (organization, department, position), and `servicecalls` count when present.
- «Мои заявки»: when `GET /api/v1/users/me` returns non-empty `servicecalls` for a found employee, the list is built from those numbers (5 per page); otherwise the previous Vuex/mock path remains.
- The large left-side prototype overview blocks were removed so the UI focuses on the actual mini app shell.
- The frontend bootstrap now auto-opens the registration screen when registration is required.
- Protected backend routes now require trusted identity from a backend bearer token; `X-User-ID` remains available only as an explicit development fallback through auth config.
- The `backend-dev` container was recreated on 2026-04-15 and now confirms both `MAX_BOT_TOKEN` and `AUTH_ACCESS_TOKEN_SECRET` inside the running process environment.
- The registration form no longer ships with fake defaults; it now prefills only trusted MAX `userId` and requires `FIO`, `Organization`, `Subdivision`, and `NamePosition` to match the real ITILIUM registration contract.
- `GET /api/v1/tickets` now uses `servicecalls` from `find_employee` and loads card summaries via `POST /list_sc`; list summaries expose `creationDate` when 1C returns it.
- `GET /api/v1/tickets/responsible` now loads numbers via `POST /list_sc_responsible` and resolves cards through `GET /find_sc`, including `creationDate` for the list UI.
- `GET /api/v1/tickets/{number}` and `POST /api/v1/tickets/search` are backed by `GET /find_sc` and map `creationDate`, `deadlineDate`, `responsibleEmployeeTitle`, `change_status`, `change_responsible`, `new_state`.
- `POST /api/v1/tickets/{number}/comments` now calls `POST /add_comment` (`multipart/form-data`, without `telegram`) and re-reads the card through `find_sc`.
- `GET /api/v1/tickets/{number}/responsibles` now calls `POST /responsibles_sc` (`multipart/form-data`) because GET returns `405` in the current 1C environment.
- Marketing request flow is now scaffolded end-to-end: backend routes `GET /api/v1/marketing/services`, `GET /api/v1/marketing/subdivisions`, `POST /api/v1/marketing/requests`, dynamic schema models (`formNumber` + `fields`), and Vue wizard step-4 renderer driven by backend schema.
- `find_employee` marketing/DAX flags are parsed with tolerance for alternate 1C key names (PascalCase, Russian labels, nested `data`/`employee`, string `Истина`/`Ложь`, numeric 1/0) via `internal/api/find_employee_permissions.go`, so `GET /api/v1/users/me` can expose `canCreateMarketingRequests` even when the raw JSON key differs from the English camelCase name.
- ITILIUM client now includes legacy marketing endpoints `listServicesMarketing`, `listSubdivisionMarketing`, `create_sc_Marketing` with resilient response parsing and demo-mode fallback schemas aligned with the current marketing forms. Marketing reference endpoints use `GET` with `id` in query (`POST` returns 405 on the current 1C publication); `listServicesMarketing` returns `КомпонентаУслуги` + `НомерФормы`, while `listSubdivisionMarketing` can return `[]` on the test contour. `create_sc_Marketing` sends multipart fields using the confirmed 1C names: `id`, `Services`, required `Subdivision`, required `ExecutionDate`, optional `files`, plus service-specific fields (`LayoutName`, `Size`, `ForWhat`, `RequiredText`, `LayoutFormats`, `ThemeEvent`, `Description`, `Budget`, `LinkToFoto`, `LinkToExamples`, `FreeText`).
- Dev/prod configs set outbound ITILIUM timeout to `50s` because ticket creation and marketing requests can wait on slow 1C processing.
- Production deployment now uses nginx HTTPS with host-mounted certificate storage from the project-local `ssl/` directory (`fullchain.pem` and `privkey.pem`), plus closed internal backend networking. Certificates are provided by the administrator rather than issued through Certbot/Let's Encrypt. Postgres is intentionally removed from the production compose because runtime code currently uses `MemoryUserRepository`, Redis cache and ITILIUM as the source of truth; migrations remain dev/tooling scaffolding only.
- `documentation/production_deployment.md` now contains a detailed Ubuntu production runbook from Docker/Compose/Loki driver installation through administrator-provided TLS certificate placement, `.env`, smoke checks, MAX connection, certificate replacement, updates, backups and troubleshooting. The production app directory is documented as `/opt/docker-shared/maxapp-invest-itilium` by default, based on the admin-provided shared Docker directory; Docker commands are written for sudo-based operation.
- Production frontend API calls use relative `/api/*` URLs by default so nginx proxies them to the backend; `VITE_PUBLIC_API_BASE_URL` remains an explicit override only for non-standard deployments.
- Pending registration profiles are no longer treated as final cache hits: `GET /api/v1/users/me` rechecks ITILIUM so users added after submitting a registration form can move from pending/registration-required to found without restarting the backend.
- Production compose now passes `LOG_LEVEL`/`LOG_FORMAT` from `.env` into the backend container, so setting `LOG_LEVEL=debug` and recreating backend enables debug logs.
- Responsible ticket lists now hydrate `find_sc` summaries concurrently with both per-card and total hydration timeouts. If a user has many responsible tickets, the backend returns a partial enriched list before HTTP `write_timeout`; not-yet-enriched items remain as fallback summaries and can still be opened individually.
- Regular ticket creation through `/create_sc` now sends only the user's description text in the `description` field. Request type, department and execution date are no longer appended to the regular ticket description; marketing-specific fields remain isolated in `/create_sc_Marketing`.
- Ticket mutations now invalidate the Redis ticket-detail cache after successful comments, status changes, responsible changes and rating confirmation. After successful `change_state_sc`, the backend returns the requested new state even when the immediate follow-up `find_sc` still reports the old state, and logs this as a stale 1C read.

## Important Decisions

- Stored profiles are treated as authoritative only when `employeeFound=true` and `registrationRequired=false`.
- The in-memory user repository now starts empty and is used mainly for local registration results.
- `GET /api/v1/users/me` falls back to a registration-required profile only when lookup returns no usable employee identity.
- `initDataUnsafe` is treated as client-side convenience only; trusted identity comes only from validated MAX `initData` on the backend.

## Open Questions

- Real `find_employee` response fields still need to be observed in logs/debugger to confirm the best mapping for `username`, `fullName`, and `department`.
- The live `/hs/Max/find_employee` currently returns `401` with body `Пользователь с таким id не найден.Необходима регистрация` for user `100245`, even though the old aiogram-based integration reportedly worked with the same credentials.
- The real MAX `initData` payload from the moderated bot still needs end-to-end verification over an HTTPS tunnel to confirm no platform-specific field differences.
- The frontend still depends on real MAX launch context for identity; when the app is opened outside MAX or with stale cached frontend assets, the UI can still show `MAX ID: не получен` until the fresh bundle and real `window.WebApp.initData` are used.
- The real HTTP response body of ITILIUM `POST /registration` is still unknown; the current integration treats any non-4xx/5xx response as success and then moves the user into `registrationPending=true`.

## Next Steps

- **E2E / регресс на тестовом контуре:** см. `documentation/e2e_test_plan.md` — варианты стека (Playwright и др.), сценарии (списки заявок, смена ответственного, комментарии, согласование) и привязка к среде.
- **Local browser dev:** see `documentation/local_development.md` — `frontend/.env.development.local` (`VITE_DEBUG_USER_ID`, optional `VITE_PUBLIC_API_BASE_URL`), Vite proxy to backend on `127.0.0.1:3000`; restart Vite after env changes. For production-like checks, use real `initData` after bot moderation / HTTPS tunnel.
- Verify `GET /api/v1/users/me` in the real environment and inspect the actual `find_employee` payload in logs/debugger.
- After bot moderation finishes, raise an HTTPS tunnel through `tuna`, point the Telegram bot to that public URL, and then continue testing through the real bot entrypoint.
- After bot moderation finishes, set `MAX_BOT_TOKEN`, open the mini app through the real MAX bot, and verify `POST /api/v1/auth/max/validate` with live `window.WebApp.initData`.
- Repeat the live check with a known registered employee id to confirm the success payload field names for `profileFromLookup(...)`.
- Compare the live request with the legacy aiogram implementation as soon as the reference code becomes available in the workspace or outside it.
- Capture real 1C payload keys for marketing services/forms (especially exact `formNumber` key names and create_sc_Marketing field naming), then tighten parsers by removing current compatibility aliases.
- If the live payload field names differ, update `profileFromLookup(...)` mapping in `internal/services/profile_service.go`.
- If the MAX production flow works, remove the last dev fallback dependency on `X-User-ID` by switching `auth.allow_debug_identity_headers` off outside local debugging.
- Submit a live registration request and inspect backend logs for the exact upstream status and response body from `POST /registration`.
- Recent accumulated changes were split into small thematic commits and pushed to `origin/main`; local empty `config.yml` was deleted.
- For new ITILIUM contract checks, prefer filtering logs by `request_id` to observe full call chains (`handler -> itilium_client -> handler`).
- For production launch, follow `documentation/production_deployment.md`: deploy under `/opt/docker-shared/maxapp-invest-itilium`, place administrator-provided cert/key under `ssl/fullchain.pem` and `ssl/privkey.pem`, fill `.env`, run `sudo docker compose up --build -d`, then verify `/healthz`, `/readyz`, MAX auth and one live ITILIUM ticket flow.
