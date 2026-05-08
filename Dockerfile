# syntax=docker/dockerfile:1.7

# 1) Build the SvelteKit frontend.
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# 2) Build the Go server with the frontend embedded.
FROM golang:1.25-alpine AS build
RUN apk add --no-cache ca-certificates git
WORKDIR /go/src/github.com/geekgonecrazy/training-log

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Replace the placeholder webfs/dist with the actual built frontend so go:embed
# picks up the real assets.
RUN rm -rf webfs/dist && mkdir -p webfs/dist
COPY --from=frontend /app/web/build/ webfs/dist/

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# 3) Minimal runtime image.
FROM scratch AS runtime

WORKDIR /usr/local/training-log

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/server ./server
# Default config — set JWT_SECRET via env at runtime.
COPY --from=build /go/src/github.com/geekgonecrazy/training-log/config.yaml.example ./config.yaml

EXPOSE 8080 9090

# Mount /usr/local/training-log/data as a volume to persist the SQLite DB.
VOLUME ["/usr/local/training-log/data"]

CMD ["./server", "--config", "config.yaml"]
