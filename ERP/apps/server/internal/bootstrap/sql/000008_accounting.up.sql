-- ─── Accounting (isolated module) ─────────────────────────────────────────────
-- These tables are populated ONLY via the internal event bus.
-- No other module creates direct FK references here.

CREATE TABLE accounts (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT    NOT NULL UNIQUE,   -- 1010 | 3110 | …
    name        TEXT    NOT NULL,
    kind        TEXT    NOT NULL CHECK (kind IN ('asset','liability','equity','income','expense')),
    parent_id   UUID    REFERENCES accounts(id),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE journal_entries (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    posted_on   DATE    NOT NULL DEFAULT CURRENT_DATE,
    memo        TEXT    NOT NULL DEFAULT '',
    source      TEXT    NOT NULL DEFAULT 'manual' CHECK (source IN ('sale','purchase','payroll','manual','tax')),
    source_ref  UUID,   -- soft reference: no FK constraint (isolated)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID    REFERENCES users(id)
);

CREATE TABLE journal_entry_lines (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id    UUID        NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id          UUID        NOT NULL REFERENCES accounts(id),
    debit               NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (debit >= 0),
    credit              NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (credit >= 0),
    memo                TEXT        NOT NULL DEFAULT '',
    CONSTRAINT chk_jel_one_side CHECK (NOT (debit > 0 AND credit > 0))
);

CREATE TABLE expense_categories (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT    NOT NULL,   -- Аренда | Коммуналка | Маркетинг
    account_id  UUID    REFERENCES accounts(id),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE expenses (
    id                      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_category_id     UUID        NOT NULL REFERENCES expense_categories(id),
    location_id             UUID        NOT NULL REFERENCES locations(id),
    amount                  NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    paid_on                 DATE        NOT NULL DEFAULT CURRENT_DATE,
    note                    TEXT        NOT NULL DEFAULT '',
    journal_entry_id        UUID        REFERENCES journal_entries(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by              UUID        REFERENCES users(id)
);

CREATE TABLE tax_records (
    id              UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID    NOT NULL REFERENCES locations(id),
    kind            TEXT    NOT NULL,   -- ИПН | СН | ОПВ | КПН
    base            NUMERIC(14,2) NOT NULL DEFAULT 0,
    amount          NUMERIC(14,2) NOT NULL DEFAULT 0,
    period_start    DATE    NOT NULL,
    period_end      DATE    NOT NULL,
    is_paid         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed standard RK chart of accounts
INSERT INTO accounts (code, name, kind) VALUES
    ('1010', 'Касса',                    'asset'),
    ('1020', 'Расчётный счёт',           'asset'),
    ('1310', 'Товары на складе',         'asset'),
    ('1410', 'Дебиторская задолженность','asset'),
    ('2010', 'Основные средства',        'asset'),
    ('3010', 'Кредиторская задолженность','liability'),
    ('3310', 'Обязательства по НДФЛ',   'liability'),
    ('3350', 'Обязательства по ОПВ',    'liability'),
    ('5010', 'Выручка от реализации',   'income'),
    ('5020', 'Прочие доходы',            'income'),
    ('7010', 'Себестоимость продаж',     'expense'),
    ('7210', 'Административные расходы', 'expense'),
    ('7410', 'Расходы на персонал',      'expense');
