-- Increase the loyalty cashback rate from 1% to 5%.
UPDATE loyalty_rules
SET params = '{"percent": 5}', updated_at = NOW()
WHERE code = 'earn_1pct';
