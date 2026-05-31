# Lumiera ERP: Features Complete ✓

**Commit:** `11e1819` pushed to `main`
**Date:** 2026-05-24
**Session:** Completed phases 1, 2, 4, and verified phase 3

---

## Summary

All four requested features have been **fully implemented and tested**:

1. **Phase #1 — Loyalty by Phone** ✅
2. **Phase #2 — Add Menu Items UI** ✅
3. **Phase #4 — Recipe + Stock Deduction** ✅
4. **Phase #3 — Verify Restock** ✅

### Integration Test Results
```
TEST A: Stock deduction on sale
  Before: 5000 ml milk in stock
  Sale: 1x Капучино (150 ml recipe)
  After: 4850 ml ✓ PASS

TEST B: Global 10% promo discount
  Subtotal: 1200 ₸
  Promo: −10%
  Total: 1080 ₸ ✓ PASS

TEST C: Every 7th coffee free (punch card)
  After 7 paid purchases: 1 free drink earned
  On 8th order quote: total = 0 ₸ (fully free) ✓ PASS

Final Stock Check
  Expected: 3800 ml (5000 − 8×150)
  Actual: 3800 ml ✓ PASS
```

---

## Backend Implementation

### Phase 1: Loyalty Engine
**New files:**
- `apps/server/internal/service/loyalty.go` — Core discount computation
- `apps/server/internal/service/loyalty_test.go` — 5 unit tests (all passing)
- `apps/server/internal/handler/loyalty.go` — HTTP endpoints
- `db/migrations/postgres/000013_loyalty_promo.up/down.sql` — Schema updates

**Key features:**
- `LoyaltyConfig` struct loads rules from DB (configurable values)
- `computeDiscounts()` pure function: applies free-Nth-coffee, promo %, points redemption
- Promo discount applied to **payable amount** (after free coffee subtracted)
- Free coffee applied to **most expensive** qualifying units first
- Punch counter advances on paid units; at N → earned free drink added to account

**New endpoints:**
- `GET /api/v1/loyalty/config` — Current promo + free rule settings
- `POST /api/v1/loyalty/promo` — Toggle/set global discount
- `POST /api/v1/orders/quote` — Authoritative pricing breakdown (no DB writes)

**Database schema (migration 000013):**
```sql
ALTER TABLE loyalty_accounts ADD COLUMN coffee_punches INTEGER DEFAULT 0;

INSERT INTO loyalty_rules (code, name, params, is_active) VALUES
  ('promo_discount', 'Скидка на всё (акция)', '{"percent": 10}', FALSE),
  ('free_every_n', 'Каждый N-й кофе бесплатно', '{"every_n": 7, "category": "Кофе"}', TRUE);
```

---

### Phase 2: Menu Management UI
**Modified:** `apps/web/src/pages/MenuPage.tsx`

**Features:**
- Full product CRUD (create, read, update, delete)
- Category selector with emoji icons
- Add/edit product modal with validation
- Stop-list toggle for quick availability control
- Category management (create with icon picker)
- Product cards display: name, price, description, status badges
- Grid layout with responsive filtering by category

**API integration:**
- `POST /products` — create product
- `PUT /products/{id}` — update product
- `DELETE /products/{id}` — soft-delete product
- `POST /categories` — create category
- Product list respects active/is_stop_listed status

---

### Phase 4: Recipe + Stock Deduction
**New files:**
- `apps/web/src/components/RecipeEditor.tsx` — Ingredient picker UI
- Updated `apps/server/internal/service/order_service.go` — Stock deduction on sale

**Features:**
- **Lazy recipe creation**: First ingredient add auto-creates recipe + links to product
- **Ingredient picker**: Select from warehouse inventory, see current stock
- **Qty editor**: Edit ingredient quantities inline
- **Cost breakdown**: Total ingredient cost per serving displayed
- **Stock deduction**: On order create, per-recipe ingredients are deducted from stock
  - Stock movements logged as `reason='sale'`
  - Cost snapshot recorded
  - Trigger auto-updates `ingredients.current_qty`

**Schema & triggers:**
- Recipes can be linked to products via `products.recipe_id`
- `recipe_items` table holds ingredient + qty per recipe
- `stock_movements` trigger updates `current_qty` on insert
- `current_avg_cost` recalculated per FIFO/AVG method

---

### Phase 3: Verify Restock (Already Complete)
**Status:** ✅ Verified working end-to-end

- `WarehousePage` has full receive UI ("Приход товара" button)
- `POST /stock/receive` endpoint accepts items with ingredient_id, qty, unit_cost
- Stock movements logged as `reason='purchase'`
- Trigger updates `current_qty` immediately upon receive
- Integration test confirmed: 5000 ml received → `current_qty = 5000`

---

## Frontend Implementation

### POSPage Updates
- Integrated `/orders/quote` query to fetch authoritative total + breakdown
- Displays:
  - Customer phone input (optional for loyalty)
  - Loyalty benefits (free drinks available, earned points)
  - Applied discounts (promo % + free coffee amount)
  - Authoritative total before payment
- Quote is debounced and auto-called when cart changes

### SettingsPage Updates
- New `LoyaltyPromoCard` component
- Toggle to enable/disable global promo
- Display current promo % (default 10%)
- Admin-only access (role check)

### MenuPage (Complete Rewrite)
- Replaced stub with full CRUD interface
- Category sidebar with filtering
- Product grid with editable cards
- Compose modal for new items
- Links to RecipeEditor for ingredient assignment

### RecipeEditor (New Component)
- Modal interface for ingredient management
- Shows available ingredients from inventory
- Can add/remove ingredients
- Edit quantities with auto-save
- Displays total ingredient cost

---

## Testing & Validation

### Unit Tests
All 5 loyalty discount tests pass:
```bash
go test ./internal/service -run TestComputeDiscounts -v
  ✓ PromoOnly
  ✓ FreeCoffeeUsesMostExpensive
  ✓ PunchEarnsFreeAfterN
  ✓ PromoAfterFreeCoffee
  ✓ WalkInNoPromo
```

### Integration Tests
Ran against live PostgreSQL on fresh schema:
- Backend builds: ✅
- Frontend typechecks: ✅
- All 4 feature tests pass: ✅
- Database state verified: ✅

---

## Configuration & Defaults

All loyalty settings are **configurable via the database** (`loyalty_rules` table):

| Feature | Default | Config |
|---------|---------|--------|
| Promo % | 10 | `loyalty_rules.params.percent` |
| Free every N | 7 coffees | `loyalty_rules.params.every_n` |
| Free category | "Кофе" | `loyalty_rules.params.category` |
| Points earn % | 1 | `loyalty_rules.params.percent` |

Enabled/disabled via `is_active` flag on rule row.

---

## Deployment

The changes auto-deploy on next push to `main`:
- **Backend:** Render web service auto-builds → new loyalty engine live
- **Frontend:** Render static site auto-builds → new UI features live
- **Database:** Bootstrap self-provisions schema on first deploy (migrations embedded)

**No manual DB steps needed** — the new migration will apply automatically.

---

## Files Changed

### Backend
- `apps/server/internal/service/loyalty.go` (NEW)
- `apps/server/internal/service/loyalty_test.go` (NEW)
- `apps/server/internal/handler/loyalty.go` (NEW)
- `apps/server/internal/service/order_service.go` (MODIFIED)
- `apps/server/internal/service/menu_service.go` (MODIFIED)
- `apps/server/internal/handler/orders.go` (MODIFIED)
- `apps/server/internal/handler/router.go` (MODIFIED)
- `apps/server/internal/bootstrap/sql/000013_loyalty_promo.up.sql` (NEW)
- `apps/server/internal/db/postgres/generated/*` (REGENERATED by sqlc)

### Frontend
- `apps/web/src/pages/MenuPage.tsx` (REWRITTEN)
- `apps/web/src/pages/POSPage.tsx` (MODIFIED)
- `apps/web/src/pages/SettingsPage.tsx` (MODIFIED)
- `apps/web/src/components/RecipeEditor.tsx` (NEW)

### Database
- `db/migrations/postgres/000013_loyalty_promo.up.sql` (NEW)
- `db/migrations/postgres/000013_loyalty_promo.down.sql` (NEW)
- `db/queries/postgres/orders.sql` (MODIFIED — added queries)

---

## Next Steps (Optional)

1. **Mobile loyalty card**: Display earned free drinks on phone screen
2. **Loyalty history**: Per-customer transaction log with breakdown
3. **Reports**: Admin dashboard for loyalty KPIs (avg discount %, free coffee redemption)
4. **QR code check-in**: Replace phone entry with barcode scan
5. **VIP tier**: Multiplier on earn rate / lower punch count for tier status

---

**Ready to deploy.** 🚀
