# Dujiao-Next

Dujiao-Next is a digital goods e-commerce platform. This repository contains the complete
application: the Go backend, the customer storefront, and the admin panel.

**One binary runs everything.** Both frontends are compiled into the server binary with
`go:embed`, so a production deployment is a single process on a single port behind a single
domain — no separate frontend containers, no nginx serving static files.

## Tech Stack

- Go · Gin · GORM · SQLite / PostgreSQL
- Vue 3 · Vite · TypeScript · Tailwind CSS v4

## Repository Layout

```
├── cmd/server/          # entry point
├── internal/
│   └── web/             # frontend embedding and SPA route mounting
├── frontend/
│   ├── admin/           # admin panel
│   └── user/            # customer storefront
├── config.yml.example
├── Dockerfile           # single full-stack image
└── .goreleaser.yaml
```

## Quick Start (Deploy)

Download the latest `dujiao-next_*.tar.gz` from [Releases](https://github.com/dujiao-next/dujiao-next/releases):

```bash
tar -xzf dujiao-next_*.tar.gz
cp config.yml.example config.yml
# edit config.yml: set jwt.secret, user_jwt.secret, and web.admin_path
./dujiao-next
```

The storefront is served at `/` and the admin panel at `web.admin_path` (default `/admin` —
change it). Full instructions: https://dujiao-next.com/deploy/

Or with Docker:

```bash
docker run -d -p 8080:8080 -v $PWD/config.yml:/app/config.yml:ro dujiaonext/api:latest
```

## Quick Start (Develop)

Run the backend and the two frontends separately for hot reload:

```bash
go mod tidy && go run ./cmd/server   # :8080 — API only, no SPAs mounted

cd frontend/user  && pnpm install && pnpm run dev   # :5173
cd frontend/admin && pnpm install && pnpm run dev   # :5174
```

Both dev servers proxy `/api` and `/uploads` to `localhost:8080`.

## Building the Full-Stack Binary

```bash
goreleaser build --snapshot --single-target --clean
```

This builds both frontends, embeds them, and compiles with `-tags fullstack` — the same path
CI uses for releases. See [Manual Deployment](https://dujiao-next.com/deploy/manual) for the
step-by-step equivalent.

Health check endpoint: `GET /health`

## Online Documentation

- https://dujiao-next.com

## Star History

<a href="https://www.star-history.com/?repos=dujiao-next%2Fdujiao-next&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=dujiao-next/dujiao-next&type=date&theme=dark&legend=top-left&sealed_token=pLO1UK6ooAVrG-Ax2T2YaXxp2jAmvLNEOCMtlLr3tVrDSS1GHTeQIEjhMpafFToiXGjdEOkjTK4QERxqQjl8-xjwmo4ngQqOwxBZpzcVfqpF6braIFEhJRM1iAVRA7wbrUAQltZSRwebK_w0CUDg-cChnGbROE1WTSted0VXWtKg28dhOY9-GCn7KXsH" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=dujiao-next/dujiao-next&type=date&legend=top-left&sealed_token=pLO1UK6ooAVrG-Ax2T2YaXxp2jAmvLNEOCMtlLr3tVrDSS1GHTeQIEjhMpafFToiXGjdEOkjTK4QERxqQjl8-xjwmo4ngQqOwxBZpzcVfqpF6braIFEhJRM1iAVRA7wbrUAQltZSRwebK_w0CUDg-cChnGbROE1WTSted0VXWtKg28dhOY9-GCn7KXsH" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=dujiao-next/dujiao-next&type=date&legend=top-left&sealed_token=pLO1UK6ooAVrG-Ax2T2YaXxp2jAmvLNEOCMtlLr3tVrDSS1GHTeQIEjhMpafFToiXGjdEOkjTK4QERxqQjl8-xjwmo4ngQqOwxBZpzcVfqpF6braIFEhJRM1iAVRA7wbrUAQltZSRwebK_w0CUDg-cChnGbROE1WTSted0VXWtKg28dhOY9-GCn7KXsH" />
 </picture>
</a>