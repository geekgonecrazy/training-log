# Training Log

Mobile-first PWA for tracking martial-arts practice, calisthenics, and gym work. Go backend (gRPC + grpc-gateway, SQLite) serves an embedded SvelteKit frontend.

## Quick start

```bash
cp config.yaml.example config.yaml
export JWT_SECRET=$(openssl rand -hex 32)

# 1) generate proto code
make proto

# 2) bootstrap account (one-time)
#    either flip auth.registration_open: true in config.yaml and POST /v1/auth/register
#    or use the seeder:
make seed && ./bin/seed-user

# 3) build & run
make build
./bin/server
```

Open http://localhost:8080.

## Layout

- `models/habit/v1/` — `.proto` sources and generated Go (messages + grpc-gateway)
- `config/` — yaml-loaded config struct
- `store/` — Store interface + SQLite implementation (`store.Users().Get(ctx, id)`)
- `core/` — transport-agnostic business logic (auth, progression, routine state machine)
- `controllers/` — gRPC service implementations (route handlers)
- `router/` — http.Handler with grpc-gateway mux + middleware + static embed
- `cmd/server/` — main binary
- `cmd/seed-user/` — one-shot user creation
- `web/` — SvelteKit PWA

## Dev

```bash
go run ./cmd/server          # backend on :8080
cd web && npm run dev        # vite on :5173, /v1 proxied to :8080
```

See `/Users/aaron/.claude/plans/i-want-to-create-curried-forest.md` for the full design.
