# MAX Identity Integration Plan

## Why This Plan Is Needed

The current project still uses a temporary identity mechanism:

- backend reads `X-User-ID` or `userId`
- frontend sends `userId` in requests
- the same `userId` is then passed into ITILIUM calls

This is acceptable only as a migration stub.

For real verification that the client is a real MAX user, the Mini App must:

- load the real MAX JavaScript library in HTML
- receive signed/encrypted init data from MAX runtime
- send that init data to the backend
- validate and decrypt it on the backend using the application token/secret
- extract the real MAX user identifier from the validated payload
- use that identifier as the acting user id for ITILIUM

## Current Status

What is already present:

- frontend and backend flows are ready to work with a single external user id
- profile, registration and ticket flows already pass `userId` through business logic
- boot flow docs already mention `ValidateMaxInitData`

What is not implemented yet:

- real MAX JS SDK connection in frontend HTML
- backend endpoint for MAX init data validation
- token verification / decryption logic
- trusted user context sourced from MAX payload
- replacement of temporary `X-User-ID` / `userId` trust model

## Target Architecture

### Frontend

The frontend must:

1. include the real MAX Mini App JavaScript library in the HTML entrypoint
2. read init data from the MAX runtime
3. send raw init data or token payload to backend validation endpoint
4. wait for validated user profile/bootstrap response
5. stop sending arbitrary client-defined identity as the source of truth

Allowed transitional behavior:

- frontend may still pass `userId` in payloads while the backend migration is incomplete
- but that `userId` must come from validated MAX data, not from a hardcoded or manually injected value

### Backend

The backend must:

1. expose a dedicated validation endpoint, for example `POST /api/v1/auth/max/validate`
2. receive MAX init data / token payload from frontend
3. verify signature or decrypt payload using configured MAX secret/token
4. extract the real MAX user identifier
5. place that identifier into request context for downstream handlers
6. resolve the user profile in repository and ITILIUM using that identifier

### ITILIUM

Legacy aiogram uses:

- Telegram user id for lookup

Migration target:

- use real MAX user id instead of Telegram user id
- later, if needed, remap to another ITILIUM attribute without changing frontend screens

## Concrete Work Plan

### Slice 1. Frontend MAX bootstrap

- add the real MAX JS library to frontend HTML
- create a thin frontend adapter for reading MAX init data
- define what exact payload is sent to backend validation endpoint
- add a bootstrap request before calling profile/ticket APIs

Deliverable:
- frontend can obtain init data from real MAX runtime instead of a stub

### Slice 2. Backend validation endpoint

- add route like `POST /api/v1/auth/max/validate`
- define request/response models for MAX validation
- validate/decrypt payload with configured MAX secret/token
- return normalized user identity payload

Deliverable:
- backend can confirm that the caller is a real MAX user

### Slice 3. Trusted identity middleware

- replace the temporary `X-User-ID` / `userId` trust model for protected routes
- store validated MAX user id in request context
- use the context user id in profile and ticket handlers

Deliverable:
- all user-sensitive routes use trusted identity from backend validation

### Slice 4. ITILIUM migration

- pass validated MAX user id to profile and ticket services
- use MAX user id in place of Telegram id for ITILIUM lookup
- preserve backend adapter boundary so future attribute remapping is localized

Deliverable:
- ITILIUM requests operate with MAX-derived identity

### Slice 5. Transitional cleanup

- remove or restrict manual `X-User-ID` / query `userId` fallback
- keep it only for local debug if absolutely necessary and behind explicit config
- update docs and test checklist

Deliverable:
- no accidental insecure identity bypass in test/prod mode

## Config Expectations

This integration will need MAX-specific config added explicitly to the project configuration, for example:

- MAX app public identifier
- MAX secret/token for validation or decryption
- enable/disable flag for local debug bypass

The exact variable names should be chosen when the real MAX SDK contract is wired in.

## Testing Plan After Implementation

When MAX identity integration is added, verify:

1. Mini App opens inside real MAX environment and the JS library is loaded
2. frontend receives init data from MAX
3. backend validates/decrypts token successfully
4. extracted MAX user id appears in logs/context as the acting user
5. profile lookup works for a real MAX user
6. the same user id is passed further into ITILIUM calls
7. forged `X-User-ID` or manual `userId` no longer changes the acting user in protected flows

## Important Transition Rule

Until this plan is implemented:

- any `userId` passed into ITILIUM should be treated as a temporary stand-in
- the intended meaning is already "MAX user id instead of Telegram id"
- later replacement should be done in one place, not scattered across frontend screens
