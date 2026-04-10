# Migrations And Mocks

## Migrations

### Where migrations live
- SQL migrations are stored in `migrations/`.
- Every migration must have an `.up.sql` and `.down.sql` pair.
- The number prefix must stay monotonic, for example `000002_add_ticket_audit_table.up.sql`.

### Apply migrations in dev
```bash
docker compose -f docker-compose.dev.yml --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" up
```

### Roll back the latest migration
```bash
docker compose -f docker-compose.dev.yml --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" down 1
```

### Check migration version
```bash
docker compose -f docker-compose.dev.yml --profile tools run --rm migrate \
  -path /migrations \
  -database "postgres://postgres:postgres@postgres:5432/app?sslmode=disable" version
```

### Create a new migration manually inside the migrate container
If you want to create the file names yourself from inside the container:
```bash
docker compose -f docker-compose.dev.yml --profile tools run --rm migrate \
  create -ext sql -dir /migrations -seq add_ticket_audit_table
```

If the `create` command is unavailable in your image version, create the pair manually in `migrations/` and keep the same numeric sequence format.

## Mocks

### Recommended tool
- Use `mockery` for generated mocks.
- Install it locally:
```bash
go install github.com/vektra/mockery/v2@latest
```

### Why mocks are needed
- We define interfaces near the consumer.
- Services depend on interfaces, not concrete repositories or API clients.
- This makes tests deterministic and keeps them independent from Redis, Postgres and ITILIUM.

### Example generation command
Generate a mock for an interface used by a service:
```bash
mockery --dir internal/services --name ProfileRepository --output internal/mocks --outpkg mocks
```

Another example for the ITILIUM client interface:
```bash
mockery --dir internal/services --name ItiliumClient --output internal/mocks --outpkg mocks
```

### Suggested testing workflow
1. Create or update an interface near the service that consumes it.
2. Generate a mock with `mockery`.
3. Write a focused service test for:
   - successful scenario
   - validation errors
   - external dependency failure
   - edge case relevant to the business rule
4. Run:
```bash
go test ./...
```

### When a handwritten stub is still fine
- If the interface is tiny and the test is local to one file, a handwritten stub is acceptable.
- This project already shows that style in `internal/services/*_test.go`.
- For wider reuse, prefer generated mocks.
