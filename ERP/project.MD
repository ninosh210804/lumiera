````markdown
# ЗАДАЧА: Разработка ERP-системы для кофейни (Казахстан)

Ты — senior full-stack разработчик. Твоя задача — спроектировать и реализовать
ERP-систему для кофейни с поддержкой офлайн-режима, мультиплатформенной установкой
и полным циклом учёта от продажи до аналитики. Действуй итеративно: сначала
архитектура и схема БД, затем модули по приоритету. Каждый этап подтверждай
с владельцем перед реализацией.

---

## 🎯 ФУНКЦИОНАЛЬНЫЕ ТРЕБОВАНИЯ

### 1. Складской учёт и расчёт себестоимости

- Двухуровневая модель: **Ингредиенты** (зерно, молоко, стаканы, крышки, сиропы)
  и **Товары** (готовые позиции меню).
- Поддержка **полуфабрикатов** (самодельные сиропы) как промежуточный уровень.
- **Техкарты (рецепты)**: товар = N ингредиентов с количеством и единицами измерения.
- **Учёт партий поставок** со средневзвешенной себестоимостью (AVG по умолчанию,
  FIFO опционально).
- **Срок годности по партиям** (FEFO) с подсветкой скоропорта.
- **Списания по причинам**: продажа, брак, пролив, угощение, сотрудник,
  инвентаризация.
- **Приходные накладные**: фиксация поставок, автоматический пересчёт себестоимости.
- **Инвентаризация**: сверка фактических остатков с системными, фиксация
  излишков/недостач.
- **Стоп-лист**: при нулевом остатке критичного ингредиента позиция автоматически
  блокируется в меню кассы.
- **Уведомления о критических остатках** — push в интерфейс администратора.

### 2. Автоматический расчёт стоимости на кассе

- При выборе позиции бариста видит итоговую цену с учётом:
  - базовой цены товара
  - модификаторов (объём 250/350/450 мл, тип молока, сиропы, доп. шоты)
  - скидок (процент, фикс. сумма, акция «каждый 6-й кофе бесплатно»)
  - бонусных баллов из CRM (списание)
- Администратор видит маржинальность позиции в реальном времени — бариста не видит.
- **Split payment**: Kaspi QR + наличные + бонусы в одном чеке.
- **Отмена и возврат чека** — только с ПИН менеджера + запись в журнал аудита.
- **Архитектурный принцип**: продажа (факт оплаты) и списание со склада —
  два отдельных события с временными метками. Никаких DB-триггеров на остатки —
  только явные транзакции через сервисный слой (event sourcing).

### 3. Аналитика и дашборд

- **Финансовые метрики**: выручка, маржинальная прибыль, средний чек, количество
  чеков (день / неделя / месяц / произвольный период).
- **ABC-анализ товаров**: топ продаж, неликвиды, самые прибыльные модификаторы.
- **Foodcost % по позициям**.
- **Часовая тепловая карта продаж** для планирования смен.
- **Сравнение бариста**: выручка, средний чек, % допродаж.
- **Прогноз закупок** на основе средних продаж за 4 недели.
- **Экспорт в CSV/Excel** — SheetJS на фронтенде, бесплатен.
- **Дашборд адаптивный** — планшет, телефон, ноутбук.

### 4. Офлайн-режим (КРИТИЧНО)

- Касса работает **полностью без интернета**: приём заказов, расчёт, печать,
  списание с локального склада.
- Локальное хранилище: **SQLite** внутри Tauri-приложения.
- **Event sourcing**: каждая операция — событие с `client_uuid` (UUIDv4),
  временной меткой устройства и порядковым номером (`sequence`).
- **Синхронизация при восстановлении сети**:
  1. Клиент пушит накопленные события на Go-сервер.
  2. Сервер проверяет идемпотентность по `client_uuid` — дубли игнорируются.
  3. Сервер разрешает конфликты остатков (отрицательные допускаются с флагом
     `needs_review` — бариста не блокируется, администратор видит алерт).
  4. Клиент пулит обновления (цены, меню, остатки).
- **Индикатор в UI**: 🟢 Online / 🔴 Offline / 🔄 Синхронизируется /
  ⚠️ Несинхронизировано (N событий).
- **Защита от дублей**: идемпотентность по `client_uuid` на всех Go-эндпоинтах.

### 5. Схема базы данных (PostgreSQL сервер + SQLite клиент)

Все бизнес-таблицы содержат обязательные поля:
`tenant_id`, `location_id`, `created_at`, `updated_at`, `deleted_at` (soft delete),
`created_by`, `client_uuid`.

```
[Организация]
tenants (1) ──< locations (1) ──< users
users >──< roles >──< permissions

[Меню]
categories (1) ──< products (1) ──< product_modifier_groups (1) ──< modifier_options
products (1) ──< price_history

[Склад]
ingredients (1) ──< stock_batches ──> suppliers
purchase_orders (1) ──< purchase_order_items ──> ingredients
stock_movements ──> ingredients, orders, inventory_counts
inventory_counts (1) ──< inventory_count_items ──> ingredients

[Рецепты]
recipes (1) ──< recipe_items ──> ingredients
recipe_items ──> recipes   [самоссылка для полуфабрикатов]
products (1) ──── recipes

[Продажи / Касса]
orders ──> users (бариста), customers, locations, shifts
orders (1) ──< order_items ──> products
order_items (1) ──< order_item_modifiers ──> modifier_options
orders (1) ──< payments ──> payment_methods
orders (1) ──< refunds

[Клиенты / CRM]
customers (1) ──── loyalty_accounts
loyalty_accounts (1) ──< loyalty_transactions ──> orders
loyalty_rules

[Смены / Персонал]
shifts ──> users, locations
time_entries ──> users, shifts
cash_drawer_operations ──> shifts, users
payroll_records ──> users, shifts

[Бухгалтерия — изолированный модуль]
accounts
journal_entries (1) ──< journal_entry_lines ──> accounts
expense_categories (1) ──< expenses
tax_records

[Синхронизация]
sync_events [client_uuid UNIQUE, sequence, synced_at, device_id]
sync_conflicts
device_registry

[Аудит]
audit_log [user_id, action, entity, old_value, new_value, device_id, created_at]
```

**Индексы для горячих запросов:**
- `orders(location_id, created_at DESC)`, `orders(customer_id)`, `orders(shift_id)`
- `stock_movements(ingredient_id, created_at DESC)`
- `sync_events(client_uuid)` UNIQUE, `sync_events(device_id, sequence)`
- `audit_log(entity, entity_id)`, `audit_log(user_id, created_at)`

**Миграции**: **golang-migrate** (`migrate` CLI + `.sql` файлы).
Схема PostgreSQL и SQLite поддерживаются отдельными наборами миграций
в `db/migrations/postgres/` и `db/migrations/sqlite/`.

**ORM/Query builder**: **sqlc** — генерирует типобезопасный Go-код из чистого SQL.
Никакого GORM — только sqlc + ручные транзакции там где нужна явная логика.

**Row Level Security** в PostgreSQL: RLS-политики по `tenant_id` на всех
бизнес-таблицах. Go-сервер устанавливает `SET LOCAL app.tenant_id = $1`
в начале каждой транзакции.

### 6. Бухгалтерский учёт (изолированный необязательный модуль)

> **Принцип**: модуль существует и доступен администратору, но полностью изолирован.
> Ни один другой модуль не импортирует из пакета `internal/accounting`.
> Данные попадают туда только через внутренний event bus (pub/sub),
> который включается настройкой `accounting.enabled = true` в конфиге.
> Если выключен — система работает без изменений.

- План счетов, настраиваемый под стандарты РК.
- Двойная запись: продажа, закупка, ЗП, аренда, прочие расходы.
- Ручной ввод расходов (аренда, коммуналка, прочее).
- Отчёты P&L, кэш-флоу, оборотно-сальдовая ведомость — только просмотр.

### 7. Установка без интернета — с флешки или кабелем

**Никаких магазинов приложений. Только локальные установщики.**

#### Десктоп (ноутбук администратора / касса на ПК):
- **Tauri 2.0** с Go-бэкендом как sidecar-процессом (Go бинарник + SQLite
  запускаются вместе с Tauri).
- Артефакты: `.exe` (NSIS installer, Windows), `.dmg` (macOS), `.AppImage` (Linux).
- Установка: скопировать установщик на флешку → запустить на целевом ПК →
  готово, интернет не нужен.
- Для Windows: инструкция обхода SmartScreen («Подробнее → Выполнить»)
  в `docs/installation/windows.md`.

#### Планшет (Android, касса бариста):
- Тот же React-фронтенд, обёрнут **Capacitor 6**.
- Артефакт: `.apk` (не `.aab` — Play Store не нужен).
- **Установка через кабель (ADB)** с ноутбука:
  ```bash
  # Включить «Отладку по USB» на планшете (Настройки → Для разработчиков)
  adb install -r coffeeshop-pos.apk
  ```
- **Установка через флешку + OTG-переходник**:
  1. Скопировать `.apk` на флешку.
  2. Подключить флешку к планшету через USB-C OTG.
  3. Открыть файловый менеджер → нажать на `.apk` → установить.
  4. Одноразово разрешить «Установку из неизвестных источников».
- Пошаговая инструкция: `docs/installation/android-sideload.md`.

#### Автообновление без магазина:
- Go-сервер отдаёт эндпоинт `GET /api/updates/latest` с версией и URL нового
  артефакта.
- Tauri Updater проверяет при запуске — предлагает обновиться.
- Для Android: уведомление в UI «Доступна версия X.Y — скачать .apk».
- Без интернета: обновление через флешку по той же схеме.

### 8. Роли и права доступа

**Администратор:**
- Полный доступ ко всем модулям.
- Управление пользователями, ролями, настройками.
- Все финансовые отчёты, аналитика, бухгалтерия.
- Редактирование меню, цен, рецептов, складских настроек.
- Инвентаризация, приходные накладные.
- Просмотр и разрешение sync-конфликтов.
- Полный журнал аудита.

**Менеджер (опционально):**
- Касса, склад, смены.
- Подтверждение отмен и возвратов.
- Без доступа к системным настройкам и бухгалтерии.

**Бариста:**
- Вход по **4–6-значному PIN-коду** (без логина/пароля).
- Только касса: заказы, оплата, печать, выдача.
- Открытие/закрытие своей смены.
- Только своя статистика.
- Не видит: себестоимость, маржу, чужие смены, склад, настройки.

**Аудит-журнал** — логируются: отмена чека, возврат, изменение цены,
открытие денежного ящика, ручное изменение остатков, вход/выход, изменение прав.

---

## 🛠 ТЕХНИЧЕСКИЙ СТЕК — ТОЛЬКО БЕСПЛАТНЫЕ ТЕХНОЛОГИИ

### Backend — Go

```
go 1.22+

Веб-фреймворк:     github.com/go-chi/chi/v5        (MIT, легковесный роутер)
HTTP middleware:    chi middleware stack              (logging, recovery, cors)
БД PostgreSQL:     github.com/jackc/pgx/v5          (нативный PostgreSQL драйвер)
БД SQLite:         github.com/mattn/go-sqlite3       (cgo, для локального sidecar)
SQL codegen:       github.com/sqlc-dev/sqlc          (генерация типобезопасного Go из SQL)
Миграции:          github.com/golang-migrate/migrate  (SQL-файлы, postgres + sqlite)
Валидация:         github.com/go-playground/validator/v10
JWT:               github.com/golang-jwt/jwt/v5
Хеширование PIN:   golang.org/x/crypto/bcrypt
UUID:              github.com/google/uuid
Config:            github.com/spf13/viper            (YAML/ENV конфиг)
Логирование:       log/slog (stdlib, Go 1.21+)
Тесты:             testing (stdlib) + github.com/stretchr/testify
Hot reload (dev):  github.com/air-verse/air
```

**Структура Go-проекта** (стандарт golang-standards/project-layout):
```
apps/server/
├── cmd/
│   └── server/
│       └── main.go              # точка входа
├── internal/
│   ├── config/                  # viper конфиг
│   ├── db/                      # pgx pool, sqlc generated code
│   │   ├── postgres/
│   │   │   ├── migrations/      # .sql файлы golang-migrate
│   │   │   ├── queries/         # .sql запросы для sqlc
│   │   │   └── generated/       # авто-сгенерированный Go-код sqlc
│   │   └── sqlite/
│   │       ├── migrations/
│   │       ├── queries/
│   │       └── generated/
│   ├── handler/                 # HTTP handlers (chi)
│   │   ├── auth.go
│   │   ├── orders.go
│   │   ├── inventory.go
│   │   ├── analytics.go
│   │   └── sync.go
│   ├── service/                 # бизнес-логика
│   │   ├── order_service.go
│   │   ├── inventory_service.go
│   │   ├── sync_service.go
│   │   └── analytics_service.go
│   ├── domain/                  # доменные типы и интерфейсы
│   ├── middleware/              # JWT, RBAC, tenant, audit
│   ├── eventbus/                # внутренний pub/sub для изолированных модулей
│   └── accounting/              # ИЗОЛИРОВАННЫЙ модуль — только подписка на eventbus
├── pkg/
│   └── escpos/                  # ESC/POS генератор чеков (чистый Go)
├── sqlc.yaml                    # конфиг sqlc
└── Makefile                     # make migrate-up, make generate, make build
```

### Frontend

```
React 18 + TypeScript + Vite
TanStack Query v5        — server state, офлайн кэш
Zustand                  — локальный state (корзина, смена, пользователь)
Tailwind CSS + shadcn/ui — UI компоненты (MIT)
Recharts                 — графики (MIT)
SheetJS (xlsx)           — экспорт Excel (Apache 2.0)
React Router v6          — навигация
```

### Desktop wrapper
```
Tauri 2.0 (Rust + WebView2/WebKit)
  ├── Go-бэкенд как sidecar (отдельный бинарник, запускается Tauri)
  ├── SQLite через tauri-plugin-sql
  └── ESC/POS через системный Rust-плагин (Bluetooth/USB/LAN)
```

### Mobile wrapper
```
Capacitor 6 (Android .apk sideload)
  ├── @capacitor-community/sqlite
  └── ESC/POS через @capacitor-community/bluetooth-le
```

### База данных
```
PostgreSQL 16   — сервер (Supabase free tier)
SQLite 3.45+    — клиент (десктоп и планшет)
```

### Хостинг — полностью бесплатный

| Компонент      | Сервис              | Лимит бесплатного тарифа              |
|----------------|---------------------|---------------------------------------|
| PostgreSQL     | Supabase free       | 500 MB БД, 2 GB трафика               |
| Go API         | Railway free        | $5 кредитов/мес (~500 часов)          |
| Go API (альт.) | Fly.io free         | 3 shared VM, 160 GB трафика           |
| Admin фронтенд | Vercel free         | Unlimited статика, Edge Functions     |
| Файлы/бэкапы   | Cloudflare R2 free  | 10 GB хранилище, 0 egress fees        |

> **Рекомендация**: Fly.io для Go-бэкенда (стабильнее Railway, не засыпает
> при наличии хотя бы одного пинга) + Supabase PostgreSQL + Vercel фронтенд.

### Монорепо
```
pnpm workspaces + Turborepo (MIT)
Go-сервер управляется через Makefile внутри apps/server/
```

---

## 📐 СТРУКТУРА МОНОРЕПО

```
coffeeshop-erp/
├── apps/
│   ├── server/           # Go API (chi + sqlc + pgx)
│   ├── desktop/          # Tauri 2.0 (Go sidecar + React фронтенд)
│   ├── mobile/           # Capacitor 6 Android (.apk)
│   └── web/              # React Admin дашборд (Vercel)
├── packages/
│   ├── shared/           # Общие TypeScript типы и утилиты
│   ├── ui/               # Общие React-компоненты (shadcn/ui обёртки)
│   └── sync-client/      # TypeScript sync engine (IndexedDB + event queue)
├── db/
│   ├── migrations/
│   │   ├── postgres/     # SQL миграции для сервера
│   │   └── sqlite/       # SQL миграции для клиента
│   └── queries/
│       ├── postgres/     # .sql запросы → sqlc генерирует Go
│       └── sqlite/       # .sql запросы → sqlc генерирует Go
├── docs/
│   ├── adr/              # Architecture Decision Records
│   ├── installation/
│   │   ├── windows.md
│   │   ├── macos.md
│   │   └── android-sideload.md
│   └── schema.md         # ER-диаграмма (Mermaid)
├── scripts/
│   ├── build-installers.sh    # Сборка .exe / .dmg / .AppImage / .apk
│   └── release.sh             # Паковка релиза для флешки
├── docker-compose.yml    # Локальный PostgreSQL для разработки
├── turbo.json
└── pnpm-workspace.yaml
```

---

## ⚠️ КЛЮЧЕВЫЕ АРХИТЕКТУРНЫЕ ПРИНЦИПЫ

1. **sqlc, не GORM** — весь SQL пишется вручную в `.sql` файлах, sqlc генерирует
   типобезопасный Go. Никаких магических ORM, никаких N+1 по умолчанию.

2. **Никаких DB-триггеров для бизнес-логики** — остатки пересчитываются в
   Go сервис-слое при обработке событий синхронизации.

3. **Event sourcing** — продажа, списание, приход, инвентаризация = неизменяемые
   события в `sync_events`. Текущее состояние = проекция событий.

4. **Бухгалтерия изолирована** — пакет `internal/accounting` не импортируется
   нигде кроме `internal/eventbus`. Включается флагом в конфиге.

5. **Мультиточечность с первого дня** — `tenant_id` + `location_id` во всех
   бизнес-таблицах. RLS в PostgreSQL. Go middleware устанавливает tenant-контекст
   из JWT на каждый запрос.

6. **Go sidecar в Tauri** — для десктопа Go компилируется в отдельный бинарник,
   Tauri запускает его как дочерний процесс через `tauri-plugin-shell`.
   Фронтенд общается с ним через `localhost` внутри приложения.

7. **Простота кассового UI** — типовой заказ за 2-3 нажатия.

8. **Фискализация отсутствует** — только нефискальные чеки для бариста и гостя.

9. **Self-hosted** — `docker-compose up` поднимает весь стек локально
   (PostgreSQL + Go API). Работает без облака.

---

## 📦 СБОРКА И РАСПРОСТРАНЕНИЕ ЧЕРЕЗ ФЛЕШКУ

Скрипт `scripts/release.sh` должен:
```bash
#!/bin/bash
VERSION=$(git describe --tags --abbrev=0)

# 1. Собрать Go бинарники для всех платформ
GOOS=windows GOARCH=amd64 go build -o release/server-windows.exe ./apps/server/cmd/server
GOOS=darwin  GOARCH=arm64 go build -o release/server-macos   ./apps/server/cmd/server
GOOS=linux   GOARCH=amd64 go build -o release/server-linux   ./apps/server/cmd/server

# 2. Собрать Tauri installers (внутри включает Go sidecar)
cd apps/desktop && pnpm tauri build

# 3. Собрать Android APK
cd apps/mobile && npx cap build android --prod

# 4. Скопировать артефакты
mkdir -p release/v$VERSION
cp src-tauri/target/release/bundle/nsis/*.exe  release/v$VERSION/
cp src-tauri/target/release/bundle/dmg/*.dmg   release/v$VERSION/
cp src-tauri/target/release/bundle/appimage/*.AppImage release/v$VERSION/
cp android/app/build/outputs/apk/release/*.apk release/v$VERSION/

# 5. Создать инструкцию
cp docs/installation/ release/v$VERSION/INSTALL/

# 6. Упаковать
zip -r release-v$VERSION.zip release/v$VERSION/
echo "✅ Готово: release-v$VERSION.zip — копируй на флешку"
```

---

## 🚀 ПОРЯДОК РАБОТЫ — СТРОГО ПОЭТАПНО

**После каждого этапа останавливайся и жди подтверждения.**

### Этап 1 — Архитектура (без кода)
- [ ] Диаграмма компонентов системы
- [ ] Полная ER-диаграмма в Mermaid (все таблицы и связи)
- [ ] Описание Go sync engine и стратегии разрешения конфликтов
- [ ] Финальный стек с обоснованием каждого выбора
- [ ] Структура монорепо и Go-проекта

### Этап 2 — Каркас проекта
- [ ] Инициализация монорепо (pnpm + Turborepo)
- [ ] `go mod init` + структура `apps/server/`
- [ ] ESLint, Prettier, `golangci-lint`, Husky pre-commit
- [ ] `docker-compose.yml` (PostgreSQL 16 + pgAdmin)
- [ ] SQL миграции для всех таблиц (golang-migrate)
- [ ] sqlc конфиг + первые сгенерированные модели
- [ ] Makefile: `make migrate-up`, `make generate`, `make dev`, `make build`

### Этап 3 — MVP модули (строго по порядку)
1. Auth + RBAC + PIN-вход бариста (Go JWT middleware)
2. Меню, категории, товары, модификаторы
3. Техкарты и ингредиенты
4. Касса — корзина, расчёт, split payment, нефискальный чек
5. Складской учёт — списания по причинам, приходы, партии
6. Офлайн sync engine — SQLite клиент, Go sync эндпоинты, конфликты
7. Смены — clock-in/out, кассовая дисциплина
8. Дашборд и аналитика (materialized views в PostgreSQL)
9. Бухгалтерия (изолированный Go пакет, eventbus)
10. Tauri desktop build с Go sidecar + Android APK
11. Скрипт сборки для флешки + инструкции установки

### Этап 4 — Тесты и документация
- [ ] Go unit-тесты: order_service, inventory_service, sync_service
- [ ] Go интеграционные тесты: sync конфликты (testcontainers-go)
- [ ] E2E: сценарий продажи офлайн + синхронизация (Playwright)
- [ ] `docs/installation/windows.md`, `macos.md`, `android-sideload.md`
- [ ] Руководство администратора и бариста (PDF через pandoc)

---

Начни с **Этапа 1**. Пиши только архитектурные схемы и диаграммы, без кода.
После завершения задай уточняющие вопросы если есть, и жди подтверждения.
````