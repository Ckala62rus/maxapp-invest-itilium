# Working Context

## Current State

- Dev backend runs through `docker-compose.dev.yml` with Air; default `.air.toml` runs the binary without Delve so `:3000` works without GoLand; optional `.air.debug.toml` runs `dlv` on container `:40000`, published to host `localhost:40100` (Windows `excludedportrange` blocked `40000`); `.air.toml` uses `poll = true` for Docker Desktop bind mounts.
- Docker containers are currently started by the user and ready for the next external integration checks.
- The ITILIUM dev target is built from `ITILIUM_HOST` in `.env` through `docker-compose.dev.yml`.
- TLS verification is temporarily disabled with `itilium.insecure_skip_verify: true` because the current test host is addressed by IP.
- The legacy user lookup contract was updated to `POST /find_employee` under `/itilium-test/hs/Max/` with form field `id`.
- Live verification on 2026-04-15 confirmed `GET /api/v1/users/me` reaches ITILIUM and currently maps the upstream `401` for user `100245` into `registrationRequired=true`.
- A new Telegram bot was created by the user and is currently waiting for moderation before end-to-end mini app checks can start.
- The app now has a first MAX auth slice: frontend loads `max-web-app.js`, reads `window.WebApp.initData`, exchanges it through `POST /api/v1/auth/max/validate`, and then uses a backend bearer token for protected API calls.
- Live verification on 2026-04-15 confirmed MAX auth validation works with real `initData`; backend logs show `user_id=40367639` and then `find_employee` in ITILIUM returns `401` with `Пользователь с таким id не найден.Необходима регистрация`.

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

- **Local browser dev:** see `documentation/local_development.md` — `frontend/.env.development.local` (`VITE_DEBUG_USER_ID`, optional `VITE_PUBLIC_API_BASE_URL`), Vite proxy to backend on `127.0.0.1:3000`; restart Vite after env changes. For production-like checks, use real `initData` after bot moderation / HTTPS tunnel.
- Verify `GET /api/v1/users/me` in the real environment and inspect the actual `find_employee` payload in logs/debugger.
- After bot moderation finishes, raise an HTTPS tunnel through `tuna`, point the Telegram bot to that public URL, and then continue testing through the real bot entrypoint.
- After bot moderation finishes, set `MAX_BOT_TOKEN`, open the mini app through the real MAX bot, and verify `POST /api/v1/auth/max/validate` with live `window.WebApp.initData`.
- Repeat the live check with a known registered employee id to confirm the success payload field names for `profileFromLookup(...)`.
- Compare the live request with the legacy aiogram implementation as soon as the reference code becomes available in the workspace or outside it.
- If the live payload field names differ, update `profileFromLookup(...)` mapping in `internal/services/profile_service.go`.
- If the MAX production flow works, remove the last dev fallback dependency on `X-User-ID` by switching `auth.allow_debug_identity_headers` off outside local debugging.
- Submit a live registration request and inspect backend logs for the exact upstream status and response body from `POST /registration`.
- Recent accumulated changes were split into small thematic commits and pushed to `origin/main`; local empty `config.yml` was deleted.
