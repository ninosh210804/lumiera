-- ─── Locations ────────────────────────────────────────────────────────────────
CREATE TABLE locations (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL,
    address     TEXT        NOT NULL DEFAULT '',
    city        TEXT        NOT NULL DEFAULT 'Almaty',
    timezone    TEXT        NOT NULL DEFAULT 'Asia/Almaty',
    phone       TEXT,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    settings    JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- ─── Roles & Permissions ──────────────────────────────────────────────────────
CREATE TABLE roles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT        NOT NULL UNIQUE,   -- admin | manager | barista | custom
    name        TEXT        NOT NULL,
    is_system   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT        NOT NULL UNIQUE,   -- orders.create | menu.edit | …
    description TEXT        NOT NULL DEFAULT ''
);

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- ─── Users ────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    default_location_id UUID        NOT NULL REFERENCES locations(id),
    role_id             UUID        NOT NULL REFERENCES roles(id),
    full_name           TEXT        NOT NULL,
    email               TEXT        UNIQUE,
    pin_hash            TEXT        NOT NULL,   -- bcrypt of 4–6 digit PIN
    is_active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    created_by          UUID        REFERENCES users(id)
);

-- Seed system roles
INSERT INTO roles (code, name, is_system) VALUES
    ('admin',   'Администратор', TRUE),
    ('manager', 'Менеджер',      TRUE),
    ('barista', 'Бариста',       TRUE);

-- Seed all permissions
INSERT INTO permissions (code, description) VALUES
    ('orders.create',        'Создавать заказы'),
    ('orders.cancel',        'Отменять заказы (с PIN)'),
    ('orders.refund',        'Возвращать заказы (с PIN)'),
    ('orders.view',          'Просматривать заказы'),
    ('menu.view',            'Просматривать меню'),
    ('menu.edit',            'Редактировать меню и цены'),
    ('inventory.view',       'Просматривать склад'),
    ('inventory.edit',       'Редактировать склад'),
    ('inventory.receive',    'Принимать товар'),
    ('inventory.count',      'Инвентаризация'),
    ('inventory.writeoff',   'Списание товара'),
    ('shifts.open_close',    'Открывать/закрывать свою смену'),
    ('shifts.view_all',      'Просматривать чужие смены'),
    ('analytics.view',       'Просматривать аналитику'),
    ('crm.view',             'Просматривать клиентов'),
    ('crm.edit',             'Редактировать CRM'),
    ('users.manage',         'Управлять пользователями'),
    ('settings.manage',      'Системные настройки'),
    ('accounting.view',      'Просматривать бухгалтерию'),
    ('sync.manage',          'Управлять синхронизацией'),
    ('audit.view',           'Просматривать журнал аудита');

-- Seed role→permission mapping
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'admin';  -- admin gets all permissions

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'manager'
  AND p.code IN (
    'orders.create','orders.cancel','orders.refund','orders.view',
    'menu.view',
    'inventory.view','inventory.receive','inventory.count','inventory.writeoff',
    'shifts.open_close','shifts.view_all',
    'analytics.view',
    'crm.view','crm.edit'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'barista'
  AND p.code IN (
    'orders.create','orders.view',
    'menu.view',
    'shifts.open_close'
);
