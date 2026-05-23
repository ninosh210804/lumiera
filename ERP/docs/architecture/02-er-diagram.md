# 2. ER-диаграмма базы данных

Диаграмма разбита на **9 логических блоков** для читаемости. Все
бизнес-таблицы имеют обязательные поля:

```
tenant_id     UUID    NOT NULL  -- мультиточечность, RLS-ключ
location_id   UUID    NOT NULL  -- торговая точка
created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
deleted_at    TIMESTAMPTZ NULL          -- soft delete
created_by    UUID    NOT NULL           -- автор операции
client_uuid   UUID    NOT NULL UNIQUE    -- идемпотентность при sync
```

В диаграммах ниже эти поля опущены для краткости.

## 2.1 Организация и доступ

```mermaid
erDiagram
    TENANTS ||--o{ LOCATIONS : has
    TENANTS ||--o{ USERS : has
    LOCATIONS ||--o{ USERS : "работают в"
    USERS }o--o{ ROLES : assigned
    ROLES }o--o{ PERMISSIONS : grants
    USERS ||--o{ AUDIT_LOG : performs

    TENANTS {
        uuid id PK
        text name
        text country "default KZ"
        jsonb settings
    }
    LOCATIONS {
        uuid id PK
        uuid tenant_id FK
        text name
        text address
        text timezone "Asia/Almaty"
    }
    USERS {
        uuid id PK
        uuid tenant_id FK
        uuid default_location_id FK
        text full_name
        text email "nullable"
        text pin_hash "bcrypt 4-6 цифр"
        text role_type "admin|manager|barista"
        boolean is_active
    }
    ROLES {
        uuid id PK
        uuid tenant_id FK
        text code "admin|manager|barista|custom"
        text name
    }
    PERMISSIONS {
        uuid id PK
        text code "orders.create|menu.edit|..."
        text description
    }
```

## 2.2 Меню, модификаторы, цены

```mermaid
erDiagram
    CATEGORIES ||--o{ PRODUCTS : groups
    PRODUCTS ||--o{ PRODUCT_MODIFIER_GROUPS : has
    PRODUCT_MODIFIER_GROUPS ||--o{ MODIFIER_OPTIONS : contains
    PRODUCTS ||--o{ PRICE_HISTORY : "цена менялась"
    PRODUCTS ||--o| RECIPES : "техкарта"

    CATEGORIES {
        uuid id PK
        text name
        int sort_order
        text icon
    }
    PRODUCTS {
        uuid id PK
        uuid category_id FK
        uuid recipe_id FK "nullable"
        text name
        text sku
        numeric base_price
        boolean is_active
        boolean is_stop_listed "автостоп при нулевом остатке"
    }
    PRODUCT_MODIFIER_GROUPS {
        uuid id PK
        uuid product_id FK
        text name "Объём|Молоко|Сироп|Шот"
        text selection_type "single|multi"
        boolean required
        int min_select
        int max_select
    }
    MODIFIER_OPTIONS {
        uuid id PK
        uuid group_id FK
        text name "250 мл|Овсяное|Ваниль"
        numeric price_delta
        uuid linked_ingredient_id FK "опц.: какой ингредиент тратится"
        numeric ingredient_qty_delta "доп. расход к рецепту"
    }
    PRICE_HISTORY {
        uuid id PK
        uuid product_id FK
        numeric price
        timestamptz effective_from
        timestamptz effective_to
        uuid changed_by FK
    }
```

## 2.3 Склад, поставки, движения

```mermaid
erDiagram
    SUPPLIERS ||--o{ PURCHASE_ORDERS : supplies
    PURCHASE_ORDERS ||--o{ PURCHASE_ORDER_ITEMS : contains
    PURCHASE_ORDER_ITEMS }o--|| INGREDIENTS : "поставляет"
    INGREDIENTS ||--o{ STOCK_BATCHES : "партии на складе"
    INGREDIENTS ||--o{ STOCK_MOVEMENTS : "движения"
    STOCK_BATCHES ||--o{ STOCK_MOVEMENTS : "из какой партии"
    INVENTORY_COUNTS ||--o{ INVENTORY_COUNT_ITEMS : checks
    INVENTORY_COUNT_ITEMS }o--|| INGREDIENTS : counts
    INVENTORY_COUNTS ||--o{ STOCK_MOVEMENTS : "корректирует"

    SUPPLIERS {
        uuid id PK
        text name
        text contact
        text bin "БИН/ИИН Казахстан"
    }
    INGREDIENTS {
        uuid id PK
        text name
        text unit "g|ml|pcs"
        boolean is_perishable
        int default_shelf_life_days
        numeric min_stock_alert
        text cost_method "AVG|FIFO"
    }
    STOCK_BATCHES {
        uuid id PK
        uuid ingredient_id FK
        uuid purchase_order_item_id FK
        numeric qty_received
        numeric qty_remaining
        numeric unit_cost
        date received_at
        date expires_at "FEFO"
    }
    STOCK_MOVEMENTS {
        uuid id PK
        uuid ingredient_id FK
        uuid batch_id FK "nullable для AVG"
        numeric qty_delta "+приход, -расход"
        numeric unit_cost_snapshot
        text reason "sale|waste|spill|gift|staff|count|purchase"
        uuid order_id FK "nullable"
        uuid inventory_count_id FK "nullable"
        text note
    }
    PURCHASE_ORDERS {
        uuid id PK
        uuid supplier_id FK
        text status "draft|received|cancelled"
        date received_at
        numeric total_amount
    }
    PURCHASE_ORDER_ITEMS {
        uuid id PK
        uuid purchase_order_id FK
        uuid ingredient_id FK
        numeric qty
        numeric unit_cost
        date expires_at "nullable"
    }
    INVENTORY_COUNTS {
        uuid id PK
        text status "open|completed|cancelled"
        timestamptz started_at
        timestamptz completed_at
        uuid performed_by FK
    }
    INVENTORY_COUNT_ITEMS {
        uuid id PK
        uuid inventory_count_id FK
        uuid ingredient_id FK
        numeric expected_qty
        numeric actual_qty
        numeric variance "+излишек, -недостача"
    }
```

## 2.4 Рецепты и полуфабрикаты

```mermaid
erDiagram
    RECIPES ||--o{ RECIPE_ITEMS : contains
    RECIPE_ITEMS }o--o| INGREDIENTS : uses
    RECIPE_ITEMS }o--o| RECIPES : "self: полуфабрикат"

    RECIPES {
        uuid id PK
        text name
        text yield_unit
        numeric yield_qty
        text type "product|semi_finished"
    }
    RECIPE_ITEMS {
        uuid id PK
        uuid recipe_id FK
        uuid ingredient_id FK "взаимоисключающе"
        uuid sub_recipe_id FK "взаимоисключающе"
        numeric qty
        text unit
    }
```

**Правило целостности:** в каждом `recipe_items` ровно одно из
`ingredient_id` / `sub_recipe_id` заполнено (`CHECK` в SQL миграции).

## 2.5 Касса, заказы, оплаты, возвраты

```mermaid
erDiagram
    ORDERS ||--o{ ORDER_ITEMS : contains
    ORDERS ||--o{ PAYMENTS : "оплата(ы)"
    ORDERS ||--o{ REFUNDS : "возврат(ы)"
    ORDERS }o--|| USERS : "бариста"
    ORDERS }o--o| CUSTOMERS : "гость"
    ORDERS }o--|| SHIFTS : "в смене"
    ORDER_ITEMS ||--o{ ORDER_ITEM_MODIFIERS : has
    ORDER_ITEM_MODIFIERS }o--|| MODIFIER_OPTIONS : ref
    PAYMENT_METHODS ||--o{ PAYMENTS : type

    ORDERS {
        uuid id PK
        uuid barista_id FK
        uuid customer_id FK
        uuid shift_id FK
        text status "open|paid|refunded|cancelled"
        numeric subtotal
        numeric discount_total
        numeric loyalty_points_used
        numeric total
        numeric cost_total "себестоимость"
        text receipt_no "локальный номер"
        text cancel_reason "nullable"
        uuid cancelled_by FK "nullable"
    }
    ORDER_ITEMS {
        uuid id PK
        uuid order_id FK
        uuid product_id FK
        numeric qty
        numeric unit_price_snapshot
        numeric line_total
        numeric line_cost
    }
    ORDER_ITEM_MODIFIERS {
        uuid id PK
        uuid order_item_id FK
        uuid modifier_option_id FK
        numeric price_delta_snapshot
    }
    PAYMENT_METHODS {
        uuid id PK
        text code "cash|kaspi_qr|card|loyalty|gift"
        text name
        boolean opens_drawer
    }
    PAYMENTS {
        uuid id PK
        uuid order_id FK
        uuid payment_method_id FK
        numeric amount
        text external_ref "Kaspi QR txn id"
    }
    REFUNDS {
        uuid id PK
        uuid order_id FK
        numeric amount
        text reason
        uuid approved_by FK "менеджерский PIN"
    }
```

## 2.6 CRM и программа лояльности

```mermaid
erDiagram
    CUSTOMERS ||--|| LOYALTY_ACCOUNTS : has
    LOYALTY_ACCOUNTS ||--o{ LOYALTY_TRANSACTIONS : changes
    LOYALTY_TRANSACTIONS }o--o| ORDERS : "по заказу"
    LOYALTY_RULES ||--o{ LOYALTY_TRANSACTIONS : "по правилу"

    CUSTOMERS {
        uuid id PK
        text phone UK "+7..."
        text name
        date birthday
        text source "manual|qr|app"
    }
    LOYALTY_ACCOUNTS {
        uuid id PK
        uuid customer_id FK
        numeric points_balance
        int free_drinks_left "акция 6-й бесплатно"
    }
    LOYALTY_TRANSACTIONS {
        uuid id PK
        uuid loyalty_account_id FK
        uuid order_id FK
        text kind "earn|spend|adjust|gift"
        numeric points_delta
        uuid rule_id FK
    }
    LOYALTY_RULES {
        uuid id PK
        text code "earn_1pct|every_6th_free|birthday_x2"
        jsonb params
        boolean is_active
    }
```

## 2.7 Смены и расчёты с персоналом

```mermaid
erDiagram
    SHIFTS ||--o{ ORDERS : "в смене"
    SHIFTS ||--o{ TIME_ENTRIES : "часы"
    SHIFTS ||--o{ CASH_DRAWER_OPERATIONS : "касса"
    SHIFTS ||--o{ PAYROLL_RECORDS : "начисления"
    USERS ||--o{ SHIFTS : works

    SHIFTS {
        uuid id PK
        uuid user_id FK
        uuid location_id FK
        timestamptz opened_at
        timestamptz closed_at
        numeric opening_cash
        numeric closing_cash_expected
        numeric closing_cash_actual
        numeric variance
    }
    TIME_ENTRIES {
        uuid id PK
        uuid user_id FK
        uuid shift_id FK
        timestamptz clock_in
        timestamptz clock_out
        numeric hours
    }
    CASH_DRAWER_OPERATIONS {
        uuid id PK
        uuid shift_id FK
        text kind "in|out|deposit|withdrawal"
        numeric amount
        text reason
    }
    PAYROLL_RECORDS {
        uuid id PK
        uuid user_id FK
        uuid shift_id FK
        numeric hourly_rate
        numeric hours
        numeric bonus
        numeric total
    }
```

## 2.8 Бухгалтерия (изолированный модуль)

```mermaid
erDiagram
    ACCOUNTS ||--o{ JOURNAL_ENTRY_LINES : "дт/кт"
    JOURNAL_ENTRIES ||--o{ JOURNAL_ENTRY_LINES : has
    EXPENSE_CATEGORIES ||--o{ EXPENSES : groups
    EXPENSES ||--|| JOURNAL_ENTRIES : "создаёт проводку"
    TAX_RECORDS }o--o| JOURNAL_ENTRIES : ref

    ACCOUNTS {
        uuid id PK
        text code "1010|3110|..."
        text name
        text kind "asset|liability|equity|income|expense"
        uuid parent_id FK
    }
    JOURNAL_ENTRIES {
        uuid id PK
        date posted_on
        text memo
        text source "sale|purchase|payroll|manual|tax"
        uuid source_ref "id события в основном модуле"
    }
    JOURNAL_ENTRY_LINES {
        uuid id PK
        uuid journal_entry_id FK
        uuid account_id FK
        numeric debit
        numeric credit
    }
    EXPENSE_CATEGORIES {
        uuid id PK
        text name "Аренда|Коммуналка|Маркетинг"
        uuid account_id FK
    }
    EXPENSES {
        uuid id PK
        uuid expense_category_id FK
        numeric amount
        date paid_on
        text note
    }
    TAX_RECORDS {
        uuid id PK
        text kind "ИПН|СН|ОПВ|КПН"
        numeric base
        numeric amount
        date period_start
        date period_end
    }
```

**Изоляция:** ни одна из таблиц выше **не имеет внешних ключей** в
таблицы продаж/склада. Связь только через `journal_entries.source_ref`
(soft reference — UUID без FK constraint).

## 2.9 Синхронизация и аудит

```mermaid
erDiagram
    DEVICE_REGISTRY ||--o{ SYNC_EVENTS : produces
    SYNC_EVENTS ||--o{ SYNC_CONFLICTS : "может вызвать"
    USERS ||--o{ AUDIT_LOG : performs

    DEVICE_REGISTRY {
        uuid id PK
        uuid tenant_id FK
        uuid location_id FK
        text device_name
        text platform "windows|macos|linux|android"
        text app_version
        timestamptz last_seen_at
    }
    SYNC_EVENTS {
        uuid id PK
        uuid client_uuid UK "идемпотентность"
        uuid device_id FK
        bigint sequence "монотонный на устройство"
        text event_type "ORDER_CREATED|STOCK_ADJUSTED|..."
        jsonb payload
        timestamptz device_ts "время устройства"
        timestamptz server_ts "время приёма"
        text status "accepted|conflict|rejected"
    }
    SYNC_CONFLICTS {
        uuid id PK
        uuid sync_event_id FK
        text kind "negative_stock|stale_price|duplicate_uuid"
        jsonb details
        boolean needs_review
        uuid resolved_by FK
        timestamptz resolved_at
        text resolution "accept|reject|adjust"
    }
    AUDIT_LOG {
        uuid id PK
        uuid user_id FK
        text action "order.cancel|price.change|drawer.open|..."
        text entity
        uuid entity_id
        jsonb old_value
        jsonb new_value
        uuid device_id FK
        text ip
    }
```

## 2.10 Индексы для горячих запросов

```sql
-- продажи
CREATE INDEX idx_orders_location_time   ON orders (location_id, created_at DESC);
CREATE INDEX idx_orders_customer        ON orders (customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX idx_orders_shift           ON orders (shift_id);

-- склад
CREATE INDEX idx_stock_movements_ingr_t ON stock_movements (ingredient_id, created_at DESC);
CREATE INDEX idx_stock_batches_fefo     ON stock_batches (ingredient_id, expires_at)
    WHERE qty_remaining > 0;

-- синхронизация
CREATE UNIQUE INDEX uq_sync_events_uuid ON sync_events (client_uuid);
CREATE INDEX idx_sync_events_device_seq ON sync_events (device_id, sequence);
CREATE INDEX idx_sync_events_status     ON sync_events (status) WHERE status <> 'accepted';

-- аудит
CREATE INDEX idx_audit_entity           ON audit_log (entity, entity_id);
CREATE INDEX idx_audit_user_time        ON audit_log (user_id, created_at DESC);

-- мультиточечность (RLS работает быстрее, когда tenant_id первый)
CREATE INDEX idx_products_tenant        ON products (tenant_id, is_active);
CREATE INDEX idx_orders_tenant_time     ON orders (tenant_id, created_at DESC);
```

## 2.11 Materialized views для аналитики

```sql
-- Часовая тепловая карта продаж (обновляется кроном раз в час)
CREATE MATERIALIZED VIEW mv_sales_heatmap AS
SELECT
    tenant_id,
    location_id,
    date_trunc('hour', created_at) AS hour_bucket,
    extract(dow  FROM created_at) AS weekday,
    extract(hour FROM created_at) AS hour_of_day,
    count(*)                       AS orders_count,
    sum(total)                     AS revenue,
    sum(total - cost_total)        AS margin
FROM orders
WHERE status = 'paid' AND deleted_at IS NULL
GROUP BY 1, 2, 3, 4, 5;

-- ABC-анализ товаров за период (параметризован через функцию)
CREATE MATERIALIZED VIEW mv_product_abc AS
SELECT
    tenant_id,
    location_id,
    product_id,
    sum(qty)                       AS units_sold,
    sum(line_total)                AS revenue,
    sum(line_total - line_cost)    AS margin,
    sum(line_total - line_cost)
        / NULLIF(sum(line_total), 0) AS margin_pct
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE o.status = 'paid' AND o.created_at >= now() - interval '90 days'
GROUP BY 1, 2, 3;
```

## 2.12 Различия PostgreSQL vs SQLite

| Возможность | PostgreSQL (сервер) | SQLite (клиент) |
|---|---|---|
| `UUID` тип | нативный `uuid` | `TEXT` хранит UUIDv4 |
| `JSONB` | да, индексируется GIN | `TEXT` с JSON1-функциями |
| `TIMESTAMPTZ` | нативный | `TEXT` ISO-8601 UTC |
| `NUMERIC(12,4)` | нативный | `REAL` (достаточно для KZT) |
| RLS | да, `tenant_id` обязательно | нет — на клиенте только свой tenant |
| Materialized views | да | нет — считаем агрегаты на сервере и пуллим |
| Триггеры на бизнес-логике | **запрещены** | **запрещены** |

Миграции живут в двух наборах файлов, имена синхронизированы:
```
db/migrations/postgres/0001_init.up.sql
db/migrations/sqlite/0001_init.up.sql
```
Изменения схемы вносятся в **оба** файла в одном PR.

## 2.13 Открытые вопросы

1. **Себестоимость**: AVG по ингредиенту по умолчанию — подтвердить, или
   нужно выбирать FIFO для конкретных групп (зерно, молоко)?
2. **Лояльность**: одна программа на tenant, или у каждой локации своя?
3. **Налоги РК**: какие конкретно фиксируем? ИПН, СН, ОПВ, КПН — этого
   достаточно или нужны ещё (НДС если плательщик, ВОСМС)?
4. **Партии для готовой продукции** (полуфабрикаты): отслеживаем срок
   годности у самодельных сиропов или считаем «бесконечный»?
5. **Multi-currency**: KZT only или нужно закладывать USD/RUB для учёта
   импортных закупок?
