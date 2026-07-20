UPDATE loyalty_rules
SET params = '{"percent": 1}', updated_at = NOW()
WHERE code = 'earn_1pct';
