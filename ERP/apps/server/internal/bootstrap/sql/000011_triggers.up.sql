-- ─── Generic updated_at trigger ───────────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

-- Apply to all tables with updated_at
CREATE TRIGGER trg_locations_updated_at         BEFORE UPDATE ON locations          FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_roles_updated_at             BEFORE UPDATE ON roles              FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_loyalty_rules_updated_at     BEFORE UPDATE ON loyalty_rules      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_users_updated_at             BEFORE UPDATE ON users              FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_categories_updated_at        BEFORE UPDATE ON categories         FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_products_updated_at          BEFORE UPDATE ON products           FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_suppliers_updated_at         BEFORE UPDATE ON suppliers          FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_ingredients_updated_at       BEFORE UPDATE ON ingredients        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_purchase_orders_updated_at   BEFORE UPDATE ON purchase_orders    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_inventory_counts_updated_at  BEFORE UPDATE ON inventory_counts   FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_recipes_updated_at           BEFORE UPDATE ON recipes            FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_orders_updated_at            BEFORE UPDATE ON orders             FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_shifts_updated_at            BEFORE UPDATE ON shifts             FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_payroll_updated_at           BEFORE UPDATE ON payroll_records    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_accounts_updated_at          BEFORE UPDATE ON accounts           FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_tax_records_updated_at       BEFORE UPDATE ON tax_records        FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_loyalty_accounts_updated_at  BEFORE UPDATE ON loyalty_accounts   FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_customers_updated_at         BEFORE UPDATE ON customers          FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_payment_methods_updated_at   BEFORE UPDATE ON payment_methods    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─── Auto-update ingredients.current_qty & current_avg_cost on stock_movements ─
CREATE OR REPLACE FUNCTION update_ingredient_balance()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    new_qty  NUMERIC;
    new_cost NUMERIC;
BEGIN
    SELECT current_qty + NEW.qty_delta INTO new_qty
    FROM ingredients WHERE id = NEW.ingredient_id;

    -- AVG cost update only on positive movements (receipts)
    IF NEW.qty_delta > 0 THEN
        SELECT
            (i.current_avg_cost * i.current_qty + NEW.unit_cost_snapshot * NEW.qty_delta)
            / NULLIF(i.current_qty + NEW.qty_delta, 0)
        INTO new_cost
        FROM ingredients i WHERE i.id = NEW.ingredient_id;
    ELSE
        SELECT current_avg_cost INTO new_cost FROM ingredients WHERE id = NEW.ingredient_id;
    END IF;

    UPDATE ingredients
    SET current_qty      = new_qty,
        current_avg_cost = COALESCE(new_cost, current_avg_cost)
    WHERE id = NEW.ingredient_id;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_stock_movements_balance
    AFTER INSERT ON stock_movements
    FOR EACH ROW EXECUTE FUNCTION update_ingredient_balance();
