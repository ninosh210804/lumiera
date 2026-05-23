# 5. Структура монорепо и Go-проекта

## 5.1 Корень монорепо

```
coffeeshop-erp/
├── apps/
│   ├── server/                 # Go API (chi + sqlc + pgx + sqlite)
│   ├── desktop/                # Tauri 2.0 wrapper
│   ├── mobile/                 # Capacitor 6 Android
│   └── web/                    # React Admin (Vercel)
├── packages/
│   ├── shared/                 # TS типы и утилиты, общие для всех клиентов
│   ├── ui/                     # shadcn/ui обёртки + наша дизайн-система
│   └── sync-client/            # TS sync engine (outbox, push/pull, conflict UI)
├── db/
│   ├── migrations/
│   │   ├── postgres/           # SQL миграции golang-migrate (сервер)
│   │   └── sqlite/             # SQL миграции (клиент)
│   └── queries/
│       ├── postgres/           # .sql → sqlc generated Go (сервер)
│       └── sqlite/             # .sql → sqlc generated Go (sidecar/тесты)
├── docs/
│   ├── architecture/           # этап 1: то, что сейчас читаете
│   ├── adr/                    # Architecture Decision Records (по мере роста)
│   ├── installation/
│   │   ├── windows.md
│   │   ├── macos.md
│   │   ├── linux.md
│   │   └── android-sideload.md
│   ├── operations/
│   │   ├── backup.md
│   │   ├── restore.md
│   │   └── upgrade.md
│   └── schema.md               # ER-диаграмма (генерится из миграций)
├── scripts/
│   ├── build-installers.sh     # сборка .exe / .dmg / .AppImage / .apk
│   ├── release.sh              # упаковка release-vX.Y.Z.zip для флешки
│   ├── dev-up.sh               # docker-compose up + migrate + sqlc generate
│   └── seed.sh                 # демо-данные для разработки
├── .github/
│   └── workflows/
│       ├── ci.yml              # lint + test + build на каждый PR
│       └── release.yml         # сборка артефактов в R2 на тэг v*
├── docker-compose.yml          # PostgreSQL 16 + pgAdmin (dev only)
├── turbo.json                  # Turborepo pipeline
├── pnpm-workspace.yaml
├── package.json                # корневой, scripts: dev / build / lint
├── .editorconfig
├── .gitignore
├── .gitattributes              # *.go text eol=lf, *.sql text eol=lf
├── README.md                   # quickstart
└── LICENSE
```

## 5.2 Go-сервер: apps/server/

Следуем golang-standards/project-layout с упрощениями (нет `pkg/`
без необходимости — выносим только то, что реально может быть
импортировано извне).

```
apps/server/
├── cmd/
│   └── server/
│       └── main.go              # точка входа: viper config, chi router,
│                                # pgx pool, sqlite open, eventbus,
│                                # graceful shutdown
├── internal/
│   ├── config/
│   │   ├── config.go            # тип Config + viper loader
│   │   └── config_test.go
│   ├── domain/                  # доменные типы и интерфейсы (без зависимостей
│   │   ├── order.go             # на конкретные пакеты)
│   │   ├── product.go
│   │   ├── stock.go
│   │   ├── shift.go
│   │   ├── sync.go              # SyncEvent, ConflictKind, EventType
│   │   ├── tenant.go
│   │   └── errors.go            # типизированные ошибки домена
│   ├── db/
│   │   ├── postgres/
│   │   │   ├── pool.go          # pgxpool.Pool, SET LOCAL app.tenant_id
│   │   │   ├── migrations/      # 0001_init.up.sql, ...
│   │   │   ├── queries/         # *.sql → sqlc
│   │   │   └── generated/       # AUTO: sqlc generate (gitignored? см. ADR)
│   │   └── sqlite/
│   │       ├── conn.go          # sql.Open + busy_timeout + journal_mode=wal
│   │       ├── migrations/
│   │       ├── queries/
│   │       └── generated/
│   ├── handler/                 # HTTP handlers — тонкие, делегируют service
│   │   ├── auth.go              # POST /api/auth/login, /api/auth/pin
│   │   ├── orders.go            # POST /api/orders, GET /api/orders
│   │   ├── inventory.go         # POST /api/stock/receive, /writeoff, /count
│   │   ├── menu.go              # CRUD products, modifiers, prices
│   │   ├── analytics.go         # GET /api/analytics/sales, /abc, /heatmap
│   │   ├── shifts.go            # POST /api/shifts/open, /close, drawer ops
│   │   ├── sync.go              # POST /api/sync/push, GET /pull, snapshot
│   │   ├── devices.go           # POST /api/devices/heartbeat
│   │   ├── conflicts.go         # GET /api/sync/conflicts, resolve
│   │   └── router.go            # chi.Mux + mount всех routes
│   ├── service/                 # бизнес-логика, координирует db + eventbus
│   │   ├── auth_service.go
│   │   ├── order_service.go     # создаёт order + stock_movements в одной tx
│   │   ├── inventory_service.go # приходы, списания, инвентаризации
│   │   ├── menu_service.go
│   │   ├── shift_service.go
│   │   ├── sync_service.go      # apply event → проекция, fix-conflict
│   │   ├── analytics_service.go # запросы к mv_sales_heatmap и т.п.
│   │   └── pricing_service.go   # калькулятор цены с модификаторами
│   ├── middleware/
│   │   ├── jwt.go               # извлекает claims, кладёт в request context
│   │   ├── tenant.go            # SET LOCAL app.tenant_id из claims
│   │   ├── rbac.go              # проверка permissions.code
│   │   ├── audit.go             # пишет AUDIT_LOG для важных endpoints
│   │   ├── recover.go           # custom recover с логом stack
│   │   └── ratelimit.go         # 100 rpm на /api/sync/push
│   ├── eventbus/                # внутренний pub/sub
│   │   ├── bus.go               # type Bus interface { Publish, Subscribe }
│   │   ├── memory.go            # in-process реализация (channel + goroutine)
│   │   └── events.go            # типы событий шины: OrderPaid, StockReceived
│   └── accounting/              # 🔒 ИЗОЛИРОВАННЫЙ модуль
│       ├── README.md            # «никто не должен импортировать этот пакет
│       │                        #  кроме main.go и только если enabled»
│       ├── module.go            # func Register(bus, cfg) — единственная
│       │                        # публичная функция
│       ├── posting_rules.go     # бизнес → проводки (двойная запись)
│       ├── reports.go           # P&L, кэш-флоу, оборотка
│       ├── db/                  # свои миграции/queries в db/migrations/postgres/
│       └── handler.go           # /api/accounting/* — монтируется опционально
├── pkg/
│   └── escpos/                  # генератор чеков, чистый Go, никаких deps
│       ├── builder.go           # type Receipt struct + Build() []byte
│       ├── codepage.go          # CP866 / CP1251 transliteration
│       └── builder_test.go
├── test/
│   ├── integration/             # testcontainers-go: реальный Postgres
│   │   ├── sync_test.go
│   │   ├── order_test.go
│   │   └── helpers.go
│   └── fixtures/                # SQL дампы для seed-данных
├── sqlc.yaml                    # конфиг для двух БД (см. §5.4)
├── go.mod
├── go.sum
├── Makefile                     # см. §5.5
├── .air.toml                    # hot reload для разработки
├── .golangci.yml                # линтер: errcheck, gosec, gocritic, ...
└── Dockerfile                   # multi-stage build для Fly.io
```

### Правила импортов внутри `internal/`

```
cmd/server → handler → service → domain
                       ↓             ↓
                       db          eventbus
                       ↓             ↓
                       domain    (accounting подписан тут)
```

- `domain` ни от кого не зависит (кроме stdlib).
- `db` зависит только от `domain` и sqlc-generated.
- `service` зависит от `domain`, `db`, `eventbus`.
- `handler` зависит от `service`, `middleware`, `domain`.
- `accounting` зависит ТОЛЬКО от `eventbus`, `domain` и собственного
  `accounting/db`. **Никто не зависит от `accounting`** (кроме main).

Это проверяется в CI линтером `depguard`:
```yaml
depguard:
  rules:
    accounting-isolation:
      list-mode: lax
      deny:
        - pkg: "**/internal/accounting"
          desc: "accounting is isolated; only main can import it"
      files:
        - "!**/cmd/server/main.go"
        - "!**/internal/accounting/**"
```

## 5.3 Tauri-десктоп: apps/desktop/

```
apps/desktop/
├── src/                         # React UI (импортирует из packages/ui)
│   ├── App.tsx
│   ├── routes/                  # касса, склад, аналитика, настройки
│   ├── api/                     # клиент к Go sidecar (localhost:7421)
│   └── hooks/
├── src-tauri/
│   ├── Cargo.toml
│   ├── tauri.conf.json          # bundle.externalBin: ["binaries/server"]
│   ├── icons/
│   ├── binaries/                # сюда build-скрипт кладёт Go sidecar
│   │   ├── server-x86_64-pc-windows-msvc.exe
│   │   ├── server-aarch64-apple-darwin
│   │   └── server-x86_64-unknown-linux-gnu
│   └── src/
│       ├── main.rs              # стартует sidecar, передаёт port
│       └── escpos.rs            # native plugin: USB/LAN/BT принтер
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 5.4 Capacitor-мобайл: apps/mobile/

```
apps/mobile/
├── src/                         # тот же React, что и desktop, но без касс-фич
│   ├── App.tsx                  # только режим бариста
│   └── api/                     # клиент идёт напрямую в облако (нет sidecar)
├── android/                     # сгенерированный Capacitor проект (gradle)
│   └── app/build/outputs/apk/   # build артефакты
├── capacitor.config.ts
├── package.json
└── vite.config.ts
```

## 5.5 Web-админка: apps/web/

```
apps/web/
├── src/
│   ├── App.tsx
│   ├── routes/
│   │   ├── dashboard/
│   │   ├── analytics/
│   │   ├── menu/
│   │   ├── inventory/
│   │   ├── settings/
│   │   ├── accounting/          # доступна только если feature_flag.accounting
│   │   └── users/
│   ├── lib/
│   │   ├── api.ts               # axios/fetch wrapper, JWT
│   │   └── query.ts             # TanStack Query setup
│   └── components/              # импорт из packages/ui
├── package.json
├── vite.config.ts
├── vercel.json                  # deploy config
└── tsconfig.json
```

## 5.6 sqlc.yaml (apps/server/sqlc.yaml)

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/db/postgres/queries"
    schema: "internal/db/postgres/migrations"
    gen:
      go:
        package: "pgdb"
        out: "internal/db/postgres/generated"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_db_tags: true
        emit_interface: true
        emit_pointers_for_null_types: true
  - engine: "sqlite"
    queries: "internal/db/sqlite/queries"
    schema: "internal/db/sqlite/migrations"
    gen:
      go:
        package: "litedb"
        out: "internal/db/sqlite/generated"
        emit_interface: true
        emit_json_tags: true
```

## 5.7 Makefile (apps/server/Makefile)

```make
.PHONY: dev migrate-up migrate-down generate build test lint

PG_URL ?= postgres://postgres:postgres@localhost:5432/coffeeshop?sslmode=disable
SQLITE_URL ?= sqlite3://./local.db

dev:               ## hot reload локального сервера
	air -c .air.toml

migrate-up:        ## применить миграции PostgreSQL и SQLite
	migrate -path internal/db/postgres/migrations -database "$(PG_URL)" up
	migrate -path internal/db/sqlite/migrations   -database "$(SQLITE_URL)" up

migrate-down:
	migrate -path internal/db/postgres/migrations -database "$(PG_URL)" down 1
	migrate -path internal/db/sqlite/migrations   -database "$(SQLITE_URL)" down 1

generate:          ## sqlc + go generate
	sqlc generate
	go generate ./...

test:
	go test -race -coverprofile=coverage.out ./...

test-integration:
	go test -tags=integration ./test/integration/...

lint:
	golangci-lint run ./...

build:             ## кросс-компиляция для всех платформ (sidecar)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o ../desktop/src-tauri/binaries/server-x86_64-pc-windows-msvc.exe ./cmd/server
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=1 go build -o ../desktop/src-tauri/binaries/server-aarch64-apple-darwin ./cmd/server
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=1 go build -o ../desktop/src-tauri/binaries/server-x86_64-unknown-linux-gnu ./cmd/server

build-cloud:       ## для Fly.io
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/server ./cmd/server
```

## 5.8 turbo.json (корень)

```json
{
  "$schema": "https://turbo.build/schema.json",
  "tasks": {
    "build":  { "dependsOn": ["^build"], "outputs": ["dist/**", "src-tauri/target/**", "android/app/build/**"] },
    "dev":    { "cache": false, "persistent": true },
    "lint":   { "outputs": [] },
    "test":   { "dependsOn": ["^build"], "outputs": ["coverage/**"] }
  }
}
```

## 5.9 docker-compose.yml (только для dev)

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: coffeeshop
    ports: ["5432:5432"]
    volumes: ["pgdata:/var/lib/postgresql/data"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
  pgadmin:
    image: dpage/pgadmin4
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@local
      PGADMIN_DEFAULT_PASSWORD: admin
    ports: ["8080:80"]
    depends_on: [postgres]
volumes:
  pgdata:
```

## 5.10 .github/workflows/ci.yml (упрощённо)

```yaml
name: ci
on: [pull_request, push]
jobs:
  go:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env: { POSTGRES_PASSWORD: postgres }
        ports: ["5432:5432"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22" }
      - run: cd apps/server && go test -race ./...
      - run: cd apps/server && golangci-lint run
  js:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
        with: { version: 9 }
      - run: pnpm install
      - run: pnpm -r build
      - run: pnpm -r lint
      - run: pnpm -r test
```

## 5.11 Конвенции коммитов и PR

- **Conventional Commits**: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`.
- Один PR — одна логическая единица.
- ADR создаётся для решений, которые сложно откатить (выбор сервиса
  хостинга, смена БД, отказ от изоляции бухгалтерии и т.п.). Шаблон —
  `docs/adr/0000-template.md`.

## 5.12 Что лежит / не лежит в git

| Лежит | Не лежит (gitignore) |
|---|---|
| `.sql` миграции и запросы | `internal/db/*/generated/` (предмет ADR) |
| `tauri.conf.json` | `src-tauri/target/`, `binaries/` |
| `package.json`, `pnpm-lock.yaml` | `node_modules/` |
| `Makefile`, `.air.toml` | `tmp/`, `*.log`, `*.db` |
| `docker-compose.yml` | `.env`, `.env.local` |
| `.example.env` (шаблон) | `coverage.out`, `*.test` |

> **ADR-кандидат:** хранить ли sqlc-generated код в git? Pro:
> воспроизводимость билда без `sqlc` в PATH. Contra: шум в diff.
> По умолчанию **не храним** (генерим в CI), но обсуждаемо.

## 5.13 Открытые вопросы

1. **Имя репозитория и название продукта** — `coffeeshop-erp` ок,
   или хотите что-то брендированное?
2. **GitHub vs GitLab vs Gitea (self-hosted)** для git-хостинга?
3. **Лицензия проекта**: MIT / Apache-2.0 / proprietary?
4. **Одна команда из 1 чел или планируется делегация работы?** От
   этого зависит строгость code review и количество ADR.
5. **Хранить ли sqlc-generated код в git** (см. §5.12)?
