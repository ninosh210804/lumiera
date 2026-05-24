-- Punch-card counter for "every Nth coffee free" (counts paid qualifying drinks).
ALTER TABLE loyalty_accounts
    ADD COLUMN IF NOT EXISTS coffee_punches INTEGER NOT NULL DEFAULT 0;

-- Configurable rules read by the order pricing engine.
--   promo_discount: global percent discount, gated by is_active (the "10% on all" toggle)
--   free_every_n:   give one free drink for every N paid drinks in `category`
INSERT INTO loyalty_rules (code, name, params, is_active) VALUES
    ('promo_discount', 'Скидка на всё (акция)',        '{"percent": 10}',                 FALSE),
    ('free_every_n',   'Каждый N-й кофе бесплатно',     '{"every_n": 7, "category": "Кофе"}', TRUE)
ON CONFLICT (code) DO NOTHING;

-- Retire the old hard-coded 6th-free rule in favour of the configurable one.
UPDATE loyalty_rules SET is_active = FALSE WHERE code = 'every_6th_free';
