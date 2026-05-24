-- ─── Device registry ──────────────────────────────────────────────────────────
CREATE TABLE device_registry (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id     UUID        NOT NULL REFERENCES locations(id),
    user_id         UUID        REFERENCES users(id),
    device_name     TEXT        NOT NULL,
    platform        TEXT        NOT NULL DEFAULT 'unknown' CHECK (platform IN ('windows','macos','linux','android','web','unknown')),
    app_version     TEXT        NOT NULL DEFAULT '0.0.0',
    last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Sync events ──────────────────────────────────────────────────────────────
CREATE TABLE sync_events (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    client_uuid     UUID        NOT NULL UNIQUE,
    device_id       UUID        NOT NULL REFERENCES device_registry(id),
    sequence        BIGINT      NOT NULL,
    event_type      TEXT        NOT NULL,
    payload         JSONB       NOT NULL,
    device_ts       TIMESTAMPTZ NOT NULL,
    server_ts       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          TEXT        NOT NULL DEFAULT 'accepted' CHECK (status IN ('accepted','conflict','rejected')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (device_id, sequence)
);

-- ─── Sync conflicts ───────────────────────────────────────────────────────────
CREATE TABLE sync_conflicts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    sync_event_id   UUID        NOT NULL REFERENCES sync_events(id),
    kind            TEXT        NOT NULL,
    details         JSONB       NOT NULL DEFAULT '{}',
    needs_review    BOOLEAN     NOT NULL DEFAULT TRUE,
    resolved_by     UUID        REFERENCES users(id),
    resolved_at     TIMESTAMPTZ,
    resolution      TEXT        CHECK (resolution IN ('accept','reject','adjust')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Audit log ────────────────────────────────────────────────────────────────
CREATE TABLE audit_log (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        REFERENCES users(id),
    action      TEXT        NOT NULL,
    entity      TEXT        NOT NULL,
    entity_id   UUID,
    old_value   JSONB,
    new_value   JSONB,
    device_id   UUID        REFERENCES device_registry(id),
    ip          INET,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
