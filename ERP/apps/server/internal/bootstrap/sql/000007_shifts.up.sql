-- ─── Time entries ─────────────────────────────────────────────────────────────
CREATE TABLE time_entries (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id),
    shift_id    UUID        NOT NULL REFERENCES shifts(id),
    clock_in    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    clock_out   TIMESTAMPTZ,
    hours       NUMERIC(6,2),   -- set by app when clock_out is recorded
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Cash drawer operations ───────────────────────────────────────────────────
CREATE TABLE cash_drawer_operations (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    shift_id    UUID        NOT NULL REFERENCES shifts(id),
    kind        TEXT        NOT NULL CHECK (kind IN ('in','out','deposit','withdrawal','sale')),
    amount      NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    reason      TEXT        NOT NULL DEFAULT '',
    client_uuid UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by  UUID        REFERENCES users(id)
);

-- ─── Payroll ──────────────────────────────────────────────────────────────────
CREATE TABLE payroll_records (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id),
    shift_id        UUID        NOT NULL UNIQUE REFERENCES shifts(id),
    hourly_rate     NUMERIC(12,2) NOT NULL DEFAULT 0,
    hours           NUMERIC(6,2) NOT NULL DEFAULT 0,
    bonus           NUMERIC(12,2) NOT NULL DEFAULT 0,
    total           NUMERIC(12,2) GENERATED ALWAYS AS (hourly_rate * hours + bonus) STORED,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
