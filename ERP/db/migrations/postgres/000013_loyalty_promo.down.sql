DELETE FROM loyalty_rules WHERE code IN ('promo_discount', 'free_every_n');
UPDATE loyalty_rules SET is_active = TRUE WHERE code = 'every_6th_free';
ALTER TABLE loyalty_accounts DROP COLUMN IF EXISTS coffee_punches;
