# 4. Финальный технологический стек с обоснованием

Каждый блок: **выбор → что отвергнуто → почему именно этот выбор**.
Главные критерии: бесплатность, офлайн-первый подход, переносимость
(флешка-инсталл), минимум магии, типобезопасность.

## 4.1 Backend: Go 1.22+

**Что отвергли:**
- **Node.js / NestJS** — отличный для прототипа, но: единый бинарник
  сложнее (pkg/nexe), runtime тяжелее, SQLite cgo-биндинги от Node
  тащат много за собой при кросс-компиляции.
- **Rust (axum)** — идеален для бинарника, но скорость разработки и
  порог входа выше. Для бизнес-логики кофейни overkill.
- **Python (FastAPI)** — runtime, не бинарник, неприменимо для sidecar.
- **Java / Kotlin (Spring Boot)** — JVM, тяжёлый старт, не для флешки.

**Почему Go:**
1. **Один статический бинарник** на платформу — критично для распространения
   через флешку. `GOOS=windows GOARCH=amd64 go build` → `.exe`, готов.
2. **Sidecar в Tauri** работает идеально: Tauri запускает `server.exe`
   как дочерний процесс, общается через `localhost:7421`.
3. **cgo для SQLite** уже есть (`mattn/go-sqlite3`) и стабилен.
4. **pgx** — лучший PostgreSQL драйвер вне зависимостей.
5. **sqlc** даёт типобезопасность БД без ORM-магии (см. §4.4).
6. **stdlib для всего**: `net/http`, `log/slog`, `encoding/json`,
   `time` — без сюрпризов.

## 4.2 HTTP-роутер: chi v5

**Что отвергли:**
- **Gin** — популярен, но не использует `http.Handler` нативно
  (свой Context). Хуже совместим со стандартными middleware.
- **Echo** — то же замечание плюс «framework feel».
- **Fiber** — на fasthttp, не http.Handler, нестандартная история.
- **stdlib net/http** — в Go 1.22 уже хорош (`http.ServeMux` с
  путями), но всё ещё нет middleware-chain из коробки.

**Почему chi:**
1. **Идиоматичный** Go: всё на `http.Handler` и `http.HandlerFunc`.
2. **Лёгкий** (~5K LOC), без зависимостей.
3. **Middleware-stack** — встроенный logging, recovery, CORS, tracing.
4. **Сабруутинг** по группам (`r.Route("/api/v1", ...)`).
5. Активно поддерживается, стабильный API с 2016.

## 4.3 База данных: PostgreSQL 16 + SQLite 3.45

**Что отвергли:**
- **MySQL** — слабее JSON, нет RLS, нет materialized views без хаков.
- **MongoDB** — нет ACID на нескольких коллекциях, плохо для денег.
- **Только SQLite (Litestream/rqlite)** — заманчиво, но RLS,
  materialized views и concurrent writes — это PostgreSQL.

**Почему PostgreSQL на сервере:**
- RLS из коробки (`tenant_id` isolation).
- `JSONB` + GIN индексы для payload событий.
- `numeric(12,4)` для денег без float-ошибок.
- Materialized views для аналитики.
- Supabase даёт бесплатный hosted PostgreSQL.

**Почему SQLite на клиенте:**
- Один файл — backup тривиален (`cp`).
- Нет сетевого сервера → нет конфигурации.
- `tauri-plugin-sql` и `@capacitor-community/sqlite` — стандарт.
- Достаточно concurrency для 1 устройства (5 одновременных потоков).

## 4.4 SQL → Go: sqlc

**Что отвергли:**
- **GORM** — магия рефлексии, N+1 по дефолту, плохие типы (`interface{}`),
  миграции через CodeFirst, сложно дебажить SQL. **Явный анти-выбор
  по требованию TZ.**
- **squirrel / goqu (query builders)** — лучше GORM, но всё ещё runtime,
  строки склеиваются. Нет компайл-тайм проверки.
- **ent** — хорош, но это снова code-first ORM с DSL, своя кривая.
- **Чистый pgx без обвязки** — повторы кода для каждого запроса.

**Почему sqlc:**
1. **Пишем SQL, получаем Go** — никакой DSL, валидируется самим
   PostgreSQL/SQLite.
2. **Полная типобезопасность**: параметры и возвращаемые поля
   генерируются как Go-структуры.
3. **Один и тот же `.sql` файл** работает для PostgreSQL и SQLite —
   но генерируем в **два разных пакета** (`db/postgres/generated`,
   `db/sqlite/generated`), потому что некоторые запросы расходятся
   (JSONB на сервере → JSON1 на клиенте).
4. **Нет рантайма**: всё происходит в `go generate`. Production
   бинарник чист.

## 4.5 Миграции: golang-migrate

**Что отвергли:**
- **goose** — близкий конкурент, но `migrate` поддерживает SQLite
  лучше из коробки.
- **atlas** — мощно, но overkill (declarative schema, дорогой
  cloud для diff).
- **самописные** — ненужный велосипед.

**Почему migrate:**
1. **CLI + Go library** — оба режима, без зависимостей.
2. **SQL-файлы** (`0001_init.up.sql` / `0001_init.down.sql`) —
   читаются глазами, ревьювятся в PR.
3. Поддержка PostgreSQL и SQLite в одном инструменте.
4. Идемпотентный: хранит `schema_migrations` таблицу.

## 4.6 Десктоп: Tauri 2.0

**Что отвергли:**
- **Electron** — 200 МБ Chromium на каждую установку, ест RAM,
  плохо для слабых ПК администраторов на точках.
- **Wails (Go + WebView)** — Tauri зрелее, лучше документация,
  лучше sidecar поддержка через `tauri-plugin-shell`.
- **Native (Qt / GTK)** — переписывать UI с React → много работы.

**Почему Tauri 2.0:**
1. **Тонкий**: использует системный WebView (WebView2 / WebKit),
   ~10 МБ установщик.
2. **Sidecar Go**: `tauri.conf.json` → `bundle.externalBin: ["server"]`
   запускает наш Go-бинарник как дочерний процесс.
3. **Кросс-платформенно**: `.exe` / `.dmg` / `.AppImage` — три цели,
   одна кодовая база.
4. **Updater**: встроенный механизм автообновлений (опц., с подписью).
5. **Rust под капотом** — безопаснее, чем Electron native modules.

**Цена:** WebView2 на Windows может потребовать ручной установки
на старых машинах без Edge. Решение: bootstrap WebView2 redistributable
включить в инсталлятор (Tauri умеет это).

## 4.7 Мобайл: Capacitor 6

**Что отвергли:**
- **React Native** — другой UI-слой, дублирование кода.
- **Flutter** — другой язык (Dart), весь UI переписывать.
- **PWA** — нет полноценного офлайн-стэка для SQLite в браузере
  Android без хаков; Bluetooth-LE на печать — кривой через Web BT.

**Почему Capacitor 6:**
1. **Тот же React-бандл** работает в браузере, Tauri и Capacitor.
2. **`@capacitor-community/sqlite`** даёт нативный SQLite на Android.
3. **`@capacitor-community/bluetooth-le`** — ESC/POS принтеры по BLE.
4. **APK sideload**: `npx cap build android --prod` → `.apk` готов
   для копирования на флешку.
5. Активно поддерживается Ionic-командой.

## 4.8 Frontend: React 18 + TS + Vite

**Что отвергли:**
- **Vue / Svelte / Solid** — отличны, но React + TS — это найм,
  экосистема (shadcn, recharts, TanStack — все на React) и
  совместимость с Capacitor / Tauri-туториалами.
- **Next.js** — серверный рендеринг не нужен для админки и кассы
  (SPA-приложения).
- **Webpack / CRA** — Vite быстрее на dev, проще на конфиге.

**Почему React + Vite:**
1. **shadcn/ui** — копируемые компоненты (нет рантайм-зависимости).
2. **TanStack Query** — лучший клиент для server state и кэширования
   офлайн.
3. **Vite HMR** мгновенный, отличная DX.
4. **React работает в Tauri и Capacitor** без переделок.

## 4.9 State management: Zustand + TanStack Query

**Что отвергли:**
- **Redux Toolkit** — мощно, но многословно, не нужен для нашей
  сложности.
- **Jotai / Recoil** — атомарный подход, но Zustand проще и популярнее.
- **MobX** — магия прокси, плохо с TS.

**Зачем оба:**
- **TanStack Query** — всё, что приходит с сервера (меню, заказы,
  аналитика). Кэш, ретраи, оптимистичные апдейты.
- **Zustand** — локальный UI state: текущая корзина, открытая смена,
  выбранный модификатор, активный экран.

## 4.10 UI: Tailwind + shadcn/ui + Recharts

**Что отвергли:**
- **Material UI / Ant Design / Chakra** — runtime CSS-in-JS,
  тяжёлые бандлы, корпоративный look.
- **Bootstrap** — устарел.
- **Headless UI без shadcn** — выпилит много велосипедов.

**Почему этот набор:**
- Tailwind = utility-CSS, нет run-time, мелкий бандл.
- shadcn/ui = копируем компоненты под себя, нет vendor lock-in.
- Recharts = чисто-SVG чарты, MIT, без D3-сложности.

## 4.11 Excel: SheetJS (xlsx)

Бесплатная Community Edition покрывает CSV/XLSX-экспорт. Pro
не нужен. Альтернатива (ExcelJS) — нормальный fallback, но SheetJS
быстрее и меньше.

## 4.12 Принтер ESC/POS

**Чистый Go-пакет в `pkg/escpos`** генерирует байты команд:
- Заголовок: `ESC @` (init), `ESC ! 0x30` (двойной размер).
- Текст в CP866 / CP1251 для кириллицы.
- Открыть денежный ящик: `ESC p 0 25 250`.
- Резать: `GS V 0`.

Транспорт:
- **Tauri**: системный Rust-плагин общается с принтером по USB /
  LAN / Bluetooth (через `escpos-rs` или прямой socket).
- **Capacitor**: `@capacitor-community/bluetooth-le` — пишем байты
  в characteristic принтера (Goojprt PT-210 и аналоги).

## 4.13 Хостинг — все сервисы бесплатные

| Компонент | Сервис | Лимит | Почему |
|---|---|---|---|
| PostgreSQL | **Supabase free** | 500 МБ БД, 2 ГБ трафика, паузит при 0 запросов > 7д | Hosted Postgres + Studio + RLS из коробки |
| Go API | **Fly.io free** | 3 shared-CPU VM, 160 ГБ исх. трафика | Не засыпает при пинге, регион Frankfurt близко к KZ |
| Web admin | **Vercel free** | Unlimited static, 100 ГБ трафик | Идеально для SPA, авто-деплой из git |
| Артефакты | **Cloudflare R2 free** | 10 ГБ хранилище, 0 egress | Хранение .exe/.dmg/.apk для авто-апдейтов |
| Бэкапы | **Cloudflare R2** | (тот же тариф) | `pg_dump` → ssh-rclone раз в сутки |

**Альтернатива «всё локально»:** docker-compose с PostgreSQL +
Go-сервером на ноутбуке администратора, без облака. Описано
в `01-components.md § 1.5`.

## 4.14 Тестирование

| Слой | Инструмент | Почему |
|---|---|---|
| Go unit | `testing` + `testify` | stdlib + удобные ассерты |
| Go integration | `testcontainers-go` | поднимает реальный PostgreSQL в Docker для теста sync-конфликтов |
| Frontend unit | `vitest` | быстрее jest, нативный для Vite |
| Frontend e2e | `playwright` | бесплатный, кросс-браузерный, может ходить в Tauri-окно |
| Нагрузочные | `k6` (опц.) | если понадобится — лимит /sync/push 100 rps |

## 4.15 CI/CD

- **GitHub Actions** (бесплатные минуты на public-repo или private с
  лимитом 2000 мин/мес).
- Pipeline:
  1. `golangci-lint`, `go test`, `pnpm lint`, `pnpm test`.
  2. На push в `main` — `go build` всех таргетов, Tauri build,
     Capacitor build → артефакты в R2.
  3. Vercel — авто-деплой `apps/web` на push.

## 4.16 Сводка стека (TL;DR)

| Слой | Технология | Лицензия | Альтернатива (если что) |
|---|---|---|---|
| Backend | Go 1.22+ | BSD | Rust + axum |
| Роутер | chi v5 | MIT | stdlib net/http |
| Postgres driver | pgx v5 | MIT | lib/pq |
| SQLite driver | mattn/go-sqlite3 | MIT | modernc.org/sqlite (pure Go) |
| ORM/SQL | sqlc | MIT | pgx + ручные структуры |
| Миграции | golang-migrate | MIT | goose |
| JWT | golang-jwt/jwt v5 | MIT | — |
| UUID | google/uuid | BSD | — |
| Конфиг | spf13/viper | MIT | knadh/koanf |
| Логи | log/slog (stdlib) | BSD | zerolog |
| Тесты | testing + testify | MIT | — |
| Frontend | React 18 + TS + Vite | MIT | Vue 3 |
| State (server) | TanStack Query v5 | MIT | SWR |
| State (local) | Zustand | MIT | Jotai |
| UI | Tailwind + shadcn/ui | MIT | Mantine |
| Charts | Recharts | MIT | Visx |
| Excel | SheetJS (xlsx) | Apache 2.0 | ExcelJS |
| Desktop | Tauri 2.0 | MIT/Apache | Electron |
| Mobile | Capacitor 6 | MIT | React Native |
| Monorepo | pnpm + Turborepo | MIT | Nx |
| Hosting | Supabase + Fly + Vercel + R2 | — | self-hosted Docker Compose |

## 4.17 Открытые вопросы

1. **Подпись Windows-инсталлятора** (Code Signing): сертификат стоит
   ~$200-400/год. Без подписи — SmartScreen warning. Берём или
   принимаем UX «нажать Подробнее → Выполнить»?
2. **`mattn/go-sqlite3` vs `modernc.org/sqlite`**: первый требует cgo
   (нужен компилятор C при кросс-сборке), второй — pure Go (но
   медленнее ~30%). Для sidecar предпочтительнее `mattn`, нужно
   подтвердить готовность настроить cgo-окружение в CI.
3. **Self-hosted локальный режим** vs **Supabase облако**: какая
   стратегия по умолчанию? От этого зависит дефолт `app.yaml`.
4. **Push-уведомления админу** (например, «остаток молока < 1 л»):
   браузерные Web Push (бесплатно, требует HTTPS) или достаточно
   in-app баннера?
