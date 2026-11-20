# Rodzic Chat — Backend

> The API behind **Rodzic Chat**: a parent-first companion app for everyday family life, with a multi-tenant communication layer for kindergartens, nurseries and after-school clubs.

Rodzic Chat helps parents keep a child's life organised in one place — health diary, tasks, shopping, calendar, NFZ (Polish public healthcare) wait times, and a community help board. Facilities (kindergartens, nurseries, schools, extracurricular clubs) can plug into the same platform to talk to parents, report absences, manage events and meals, and settle payments.

This repository contains the Go REST API that powers the mobile apps (iOS / Android, Kotlin Multiplatform), the parent/director web dashboard (SvelteKit) and the marketing site.

---

## Table of contents

- [Product overview](#product-overview)
- [Features](#features)
- [Architecture](#architecture)
- [Tech stack](#tech-stack)
- [Project structure](#project-structure)
- [Getting started](#getting-started)
- [Configuration](#configuration)
- [Database & migrations](#database--migrations)
- [Operational tools](#operational-tools)
- [Deployment](#deployment)
- [Security notes](#security-notes)
- [Related repositories](#related-repositories)
- [Documentation](#documentation)

---

## Product overview

Rodzic Chat has two complementary sides:

| Side | Audience | What it offers |
|---|---|---|
| **Personal parent account** | Parents / guardians | Child profiles, health & illness diary, to-do lists with reminders, shared shopping lists, calendar, NFZ specialist search, "Szukam / Polecam" community help board, co-parent sharing |
| **Facility (organization) module** | Directors, co-directors, teachers | Group chats and announcements, absence reporting, events & trip consents, meal cancellations, registration/teacher invites, director dashboard, platform subscription and payments |

Design principles:

- **Parent-first.** A parent can use the app with no facility involved at all; facilities are an optional extension.
- **Multi-tenant by construction.** Almost every facility-related table carries an `OrganizationId`, so one facility's data never mixes with another's.
- **Privacy by default.** Child health data is visible **only** to parents linked to the child. No government ID (PESEL) or document scanning is stored.
- **Not an e-register.** Rodzic Chat deliberately sits *next to* a facility's official electronic register; it is a communication channel, not pedagogical documentation.

## Features

**Accounts & authentication**
- Email/password, Google and Sign in with Apple (native mobile token verification)
- Short-lived JWT access tokens + rotating refresh tokens tied to a per-device session
- Per-device logout and global session revocation (e.g. after password reset)
- Single-use, hashed password-reset tokens; soft-delete of accounts with a scheduled hard-delete reaper
- Roles: Parent, Teacher, Director, Co-Director (with a granular staff-permission catalog)
- Multi-facility membership for parents, with organization switching

**Family & child tools**
- Children, co-parents and parent invites (share a child with a second guardian)
- Child health card, illness diary, medication doses with push reminders, BMI-for-age (WHO tables)
- To-dos with repeats, reminders, snoozing, shared tasks and daily digests (opt-in)
- Shared shopping lists
- Calendar including co-parents' events
- NFZ integration: specialty tiles, nearby search by distance/wait time, saved places, 24h response cache
- "Szukam / Polecam" help board: questions, photos, best answer, category follows, activity feed
- Announcements by category (info / sale / event / other)

**Facility module**
- Groups with auto-created channels (parents-only chat and an announcement channel)
- Chats, direct messages, polls, read receipts, chat invites
- Absence reports, events with participation confirmation and consent-document uploads
- Meals and meal-cancellation tracking
- Parent registration links and teacher invites (link-based, revocable)
- Guardian-claim workflow: director approval when a child's birth date doesn't match, preventing impersonation
- Director activity feed and monthly summary
- Platform subscription tracking, Stripe Connect accounts and payments

**Platform**
- Push notifications (FCM) and background schedulers for reminders and dose alerts
- S3-compatible object storage (Cloudflare R2) for avatars, media and attachments
- Transactional email via SMTP
- Optional AI chat assistant (Anthropic Claude)
- Per-endpoint rate limiting on authentication and recovery routes

## Architecture

The code follows a layered, domain-oriented layout. Each domain (e.g. `children`, `todos`, `childhealth`) is split across three layers:

```
HTTP request
   │
   ▼
transport/<domain>      gorilla/mux handlers: parsing, auth context, JSON responses
   │
   ▼
internal/<domain>       business logic, authorization rules, orchestration
   │
   ▼
common/<domain>         domain models + request/response types
   │
   ▼
database/gormdb         generic GORM query layer (conditions, mutations, transactions)
   │
   ▼
PostgreSQL
```

Routes are grouped into `/public` (registration, login, password recovery, invite lookup, webhooks) and `/private` (everything behind the JWT auth middleware, with role- and permission-based guards applied per route).

## Tech stack

| Area | Technology |
|---|---|
| Language | Go 1.27 |
| HTTP router | [gorilla/mux](https://github.com/gorilla/mux) |
| ORM / DB | [GORM](https://gorm.io) + PostgreSQL (pgx) |
| Auth | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)), bcrypt, OAuth2 (Google / Apple) |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) |
| Payments | [Stripe](https://stripe.com) (Connect + webhooks) |
| Storage | S3-compatible (Cloudflare R2) via minio-go |
| Push | Firebase Cloud Messaging |
| Observability | OpenTelemetry |
| Packaging | Docker (multi-stage, Alpine) |

## Project structure

```
.
├── main.go                  # Entry point (starts the HTTP server)
├── big.migrate.go           # One-shot schema bootstrap (see "Database & migrations")
├── api_bucket/              # Object-storage upload / download / delete
├── cmd/                     # Standalone operational binaries
│   ├── reaper/              # Hard-deletes long-soft-deleted accounts
│   ├── backfillaccountorgs/ # Backfills parent ↔ organization membership
│   └── migratechat/         # Chat data migration
├── database/
│   ├── common/              # Shared DB types
│   └── gormdb/              # Generic query / mutation / condition layer
├── library/                 # Reusable helpers (crypto, pinhash, random, routines, ...)
├── logger/                  # Logging
├── project/
│   ├── common/              # Domain models and request/response types
│   └── service/
│       ├── internal/        # Business logic per domain
│       └── transport/       # HTTP handlers + route wiring (module.go)
├── variables/               # Environment-driven configuration
├── doc/schema/              # Auto-generated DB schema docs (tbls)
├── docs/                    # Architecture PDFs
├── Dockerfile
└── Makefile
```

## Getting started

### Prerequisites

- Go 1.27+
- PostgreSQL 14+
- An S3-compatible bucket (Cloudflare R2 or MinIO) — required at startup
- Stripe test keys and SMTP credentials — required at startup (see below)

### Run locally

```bash
# 1. Create a database
createdb rodzicchat

# 2. Provide configuration (see "Configuration")
cp .env.example .env   # then fill in the values

# 3. Start the API (listens on :8080 by default)
set -a; . ./.env; set +a
go run .
```

The server fails fast at startup if a required environment variable is missing.

### Build & run with Docker

```bash
docker build -t rodzic-chat-backend .
docker run --rm -p 8080:8080 --env-file .env rodzic-chat-backend
```

## Configuration

All configuration is via environment variables (see [`variables/secrets.go`](variables/secrets.go)). Secrets are never committed — `.env` is git-ignored.

**Required**

| Variable | Purpose |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Signing key for access/refresh tokens |
| `SMTP_LOGIN`, `SMTP_PASSWORD` | Transactional email credentials |
| `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET` | Stripe API and webhook verification |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Google OAuth (web flow) |
| `R2_ACCESS_KEY`, `R2_SECRET_KEY`, `R2_ENDPOINT` | Object storage |

**Optional**

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `BACKEND_URL` / `FRONTEND_URL` | localhost / platform URL | Used in links and redirects |
| `CORS_ALLOWED_ORIGINS` | — | Allowed browser origins |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_FROM` | Brevo relay | Email transport |
| `GOOGLE_OAUTH_CLIENT_ID`, `APPLE_AUDIENCE` | — | Native sign-in audience values |
| `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL` | — | Enables the AI chat assistant |
| `GOOGLE_PLACES_API_KEY` | — | Richer place search (falls back to OSM) |
| `FIELD_ENCRYPTION_KEY` | — | Optional field-level encryption |
| `PLATFORM_BANK_ACCOUNT_NUMBER`, `PLATFORM_BANK_ACCOUNT_HOLDER` | placeholders | Platform subscription transfer details |
| `ALLOW_TEST_ENDPOINTS` | `false` | Enables test-only endpoints — never in production |

## Database & migrations

The backend uses GORM with PostgreSQL. Conventions (full details in [`DATABASE.md`](DATABASE.md)):

- IDs are random UUID strings; money is stored as integer minor units (grosze), never floats.
- Enums are stored as integers — **new values are always appended, never inserted mid-list**.
- Only `Account` is soft-deleted; everything else is hard-deleted.
- Foreign keys use `ON DELETE CASCADE` so deleting an account cleans up its data.

Schema bootstrap lives in [`big.migrate.go`](big.migrate.go) (`BigMigrate()`), which creates the tables, constraints and indexes for the parent-facing app. It is invoked manually (uncomment the call in `main.go` for a one-off run, then restore it). The facility-module tables are defined in the same file and enabled via the `modules` flags.

A browsable schema reference is generated with [tbls](https://github.com/k1LoW/tbls):

```bash
tbls doc --force      # outputs to doc/schema
```

## Operational tools

| Command | Purpose |
|---|---|
| `go run ./cmd/reaper` | Permanently removes accounts soft-deleted 3+ months ago, plus orphaned data. Run on a schedule. |
| `go run ./cmd/backfillaccountorgs` | One-time backfill of parent ↔ organization membership for pre-existing accounts. |
| `go run ./cmd/migratechat` | One-time chat data migration. |

## Deployment

The service ships as a single static binary in a small Alpine image, built for `linux/amd64`, and is deployed to **DigitalOcean App Platform** from Docker Hub.

```bash
make build                       # build the image
make push DOCKERHUB_USER=<user>  # push to Docker Hub
make deploy DO_APP_ID=<app-id>   # trigger a deployment via doctl
make release                     # push + deploy
```

Schedule `cmd/reaper` (e.g. daily) as a separate job in your environment.

## Security notes

- Passwords are hashed with bcrypt (72-byte limit enforced by validation); refresh tokens and reset tokens are stored only as hashes.
- Access tokens and refresh tokens are distinct kinds; the auth middleware rejects refresh tokens on API routes.
- Auth, registration, refresh and recovery endpoints are rate limited.
- Joining an existing account to a second facility requires that account's password.
- Optional per-group child birth-date verification and a director-approved guardian-claim flow protect against impersonation.
- All required secrets come from the environment; the process refuses to start without them.

## Related repositories

Rodzic Chat spans several repositories:

| Repository | Description |
|---|---|
| `Rodzic-Chat` (this repo) | Go backend API |
| `RodzicChat` | Kotlin Multiplatform mobile app (SwiftUI iOS + Android + shared logic) |
| `Rodzic-Chat-WEB` | SvelteKit web dashboard for parents and directors |
| `Rodzic-Chat-Site` | Static marketing website |

## Documentation

- [`DATABASE.md`](DATABASE.md) — every table, its purpose and relationships
- [`doc/schema/`](doc/schema) — generated schema reference
- [`docs/`](docs) — architecture documents (backend, iOS, Android)
- [`XPLAN/`](XPLAN) — product and technical planning notes

---

© Flashy Code. All rights reserved.
