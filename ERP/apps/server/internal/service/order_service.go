package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

type OrderService struct {
	q    *pgdb.Queries
	pool *pgxpool.Pool
}

func NewOrderService(pool *pgxpool.Pool, q *pgdb.Queries) *OrderService {
	return &OrderService{q: q, pool: pool}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type OrderDTO struct {
	ID                uuid.UUID      `json:"id"`
	LocationID        uuid.UUID      `json:"location_id"`
	ShiftID           *uuid.UUID     `json:"shift_id,omitempty"`
	BaristaID         uuid.UUID      `json:"barista_id"`
	CustomerID        *uuid.UUID     `json:"customer_id,omitempty"`
	Status            string         `json:"status"`
	Subtotal          float64        `json:"subtotal"`
	DiscountTotal     float64        `json:"discount_total"`
	LoyaltyPointsUsed float64        `json:"loyalty_points_used"`
	Total             float64        `json:"total"`
	ReceiptNo         string         `json:"receipt_no"`
	ClientUUID        uuid.UUID      `json:"client_uuid"`
	CreatedAt         time.Time      `json:"created_at"`
	Items             []OrderItemDTO `json:"items,omitempty"`
	Payments          []PaymentDTO   `json:"payments,omitempty"`
	LoyaltyEarned     float64        `json:"loyalty_earned,omitempty"`
}

type OrderItemDTO struct {
	ID                uuid.UUID          `json:"id"`
	ProductID         uuid.UUID          `json:"product_id"`
	ProductName       string             `json:"product_name"`
	Qty               float64            `json:"qty"`
	UnitPriceSnapshot float64            `json:"unit_price"`
	LineTotal         float64            `json:"line_total"`
	ClientUUID        uuid.UUID          `json:"client_uuid"`
	Modifiers         []OrderModifierDTO `json:"modifiers,omitempty"`
}

type OrderModifierDTO struct {
	ID               uuid.UUID `json:"id"`
	ModifierOptionID uuid.UUID `json:"modifier_option_id"`
	OptionName       string    `json:"option_name"`
	PriceDelta       float64   `json:"price_delta"`
}

type PaymentDTO struct {
	ID          uuid.UUID `json:"id"`
	MethodCode  string    `json:"method_code"`
	MethodName  string    `json:"method_name"`
	Amount      float64   `json:"amount"`
	ExternalRef string    `json:"external_ref,omitempty"`
}

type ReceiptItemDTO struct {
	Name      string   `json:"name"`
	Qty       float64  `json:"qty"`
	UnitPrice float64  `json:"unit_price"`
	LineTotal float64  `json:"line_total"`
	Modifiers []string `json:"modifiers,omitempty"`
}

type ReceiptDTO struct {
	ReceiptNo     string           `json:"receipt_no"`
	LocationID    uuid.UUID        `json:"location_id"`
	BaristaID     uuid.UUID        `json:"barista_id"`
	CreatedAt     time.Time        `json:"created_at"`
	Items         []ReceiptItemDTO `json:"items"`
	Subtotal      float64          `json:"subtotal"`
	DiscountTotal float64          `json:"discount_total"`
	LoyaltyUsed   float64          `json:"loyalty_points_used"`
	Total         float64          `json:"total"`
	Payments      []PaymentDTO     `json:"payments"`
	LoyaltyEarned float64          `json:"loyalty_earned"`
}

type CustomerProfileDTO struct {
	Phone          string  `json:"phone"`
	CustomerFound  bool    `json:"customer_found"`
	Balance        float64 `json:"balance"`
	FreeDrinksLeft int     `json:"free_drinks_left"`
	CoffeePunches  int     `json:"coffee_punches"`
}

// ─── Inputs ───────────────────────────────────────────────────────────────────

type CreateOrderInput struct {
	LocationID         uuid.UUID
	ShiftID            *uuid.UUID
	BaristaID          uuid.UUID
	CustomerPhone      string
	LoyaltyPointsToUse float64
	ManualDiscountPct  float64 // per-order barista discount (e.g. paper ticket for 10% off)
	ClientUUID         uuid.UUID
	Items              []OrderItemInput
	Payments           []PaymentInput

	// Comp marks an order as products taken without payment (e.g. Zhandos took
	// stock off the shelf). Ingredients are still deducted from the warehouse,
	// but the order is excluded from revenue and tracked as a debt per recipient.
	Comp          bool
	CompRecipient string
}

type OrderItemInput struct {
	ProductID         uuid.UUID
	Qty               float64
	ModifierOptionIDs []uuid.UUID
	ClientUUID        uuid.UUID
}

type PaymentInput struct {
	MethodCode  string
	Amount      float64
	ExternalRef string
}

type CancelOrderInput struct {
	OrderID     uuid.UUID
	Reason      string
	RequestedBy uuid.UUID
}

// ─── CreateOrder ──────────────────────────────────────────────────────────────

func (svc *OrderService) CreateOrder(ctx context.Context, in CreateOrderInput) (*OrderDTO, error) {
	if len(in.Items) == 0 {
		return nil, domain.NewBadRequest("order must have at least one item")
	}

	tx, err := svc.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	q := svc.q.WithTx(tx)

	pmMap, err := loadPaymentMethods(ctx, q)
	if err != nil {
		return nil, err
	}

	type ingredientDeduction struct {
		ingredientID uuid.UUID
		qtyDelta     float64 // negative = consumed from stock
		costSnapshot float64
	}

	type calcItem struct {
		productID   uuid.UUID
		productName string
		qty         float64
		unitPrice   float64
		lineTotal   float64
		lineCost    float64
		clientUUID  uuid.UUID
		modIDs      []uuid.UUID
		modDeltas   []float64
		modNames    []string
		deductions  []ingredientDeduction
	}

	cfg, err := loadLoyaltyConfig(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("load loyalty config: %w", err)
	}

	var calcItems []calcItem
	var subtotal float64
	var costTotal float64
	var coffeeUnitPrices []float64

	for _, item := range in.Items {
		prod, err := q.GetProduct(ctx, pgtype.UUID{Bytes: item.ProductID, Valid: true})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.NewNotFound("product")
			}
			return nil, fmt.Errorf("get product: %w", err)
		}
		if !prod.IsActive || prod.IsStopListed {
			return nil, domain.NewBadRequest(fmt.Sprintf("product '%s' is not available", prod.Name))
		}

		unitPrice := floatFromNumeric(prod.BasePrice)
		if prod.SalePrice.Valid {
			unitPrice = floatFromNumeric(prod.SalePrice) // active sale event overrides base price
		}
		var modDeltas []float64
		var modNames []string

		for _, modID := range item.ModifierOptionIDs {
			mod, err := q.GetModifierOption(ctx, pgtype.UUID{Bytes: modID, Valid: true})
			if err != nil {
				return nil, fmt.Errorf("get modifier option: %w", err)
			}
			delta := floatFromNumeric(mod.PriceDelta)
			unitPrice += delta
			modDeltas = append(modDeltas, delta)
			modNames = append(modNames, mod.Name)
		}

		lineTotal := math.Round(unitPrice*item.Qty*100) / 100
		subtotal += lineTotal

		// Track per-unit prices of qualifying drinks for the free-Nth-coffee rule.
		if prod.CategoryName == cfg.FreeCategory {
			for n := 0; n < int(item.Qty); n++ {
				coffeeUnitPrices = append(coffeeUnitPrices, unitPrice)
			}
		}

		itemUUID := item.ClientUUID
		if itemUUID == (uuid.UUID{}) {
			itemUUID = uuid.New()
		}

		// Expand the product's recipe into ingredient stock deductions (if linked).
		var deductions []ingredientDeduction
		var lineCost float64
		if prod.RecipeID.Valid {
			recipeItems, err := q.GetRecipeItems(ctx, prod.RecipeID)
			if err != nil {
				return nil, fmt.Errorf("get recipe items: %w", err)
			}
			for _, ri := range recipeItems {
				if !ri.IngredientID.Valid {
					continue // sub-recipe expansion not handled yet; skip
				}
				perUnit := floatFromNumeric(ri.Qty)
				avgCost := floatFromNumeric(ri.CurrentAvgCost)
				// Convert recipe qty into the ingredient's storage unit so
				// stock deductions and cost match — see units.go and /3.
				ingUnit := ri.Unit
				if ri.IngredientUnit != nil {
					ingUnit = *ri.IngredientUnit
				}
				perUnitInIngUnit := convertQty(perUnit, ri.Unit, ingUnit)
				consumed := perUnitInIngUnit * item.Qty
				lineCost += consumed * avgCost
				deductions = append(deductions, ingredientDeduction{
					ingredientID: uuid.UUID(ri.IngredientID.Bytes),
					qtyDelta:     -consumed,
					costSnapshot: avgCost,
				})
			}
			lineCost = math.Round(lineCost*100) / 100
			costTotal += lineCost
		}

		calcItems = append(calcItems, calcItem{
			productID:   item.ProductID,
			productName: prod.Name,
			qty:         item.Qty,
			unitPrice:   unitPrice,
			lineTotal:   lineTotal,
			lineCost:    lineCost,
			clientUUID:  itemUUID,
			modIDs:      item.ModifierOptionIDs,
			modDeltas:   modDeltas,
			modNames:    modNames,
			deductions:  deductions,
		})
	}

	var (
		customerID       pgtype.UUID
		loyaltyAccountID pgtype.UUID
		custState        loyaltyState
	)
	// Comps carry no payment and earn no loyalty; ignore any phone/loyalty input
	// and charge full retail value so the recipient's tally reflects real cost.
	if in.CustomerPhone != "" && !in.Comp {
		custID, loyAcctID, st, err := getOrCreateCustomer(ctx, q, in.CustomerPhone)
		if err != nil {
			return nil, err
		}
		customerID = pgtype.UUID{Bytes: custID, Valid: true}
		loyaltyAccountID = pgtype.UUID{Bytes: loyAcctID, Valid: true}
		custState = st
	}

	disc := computeDiscounts(subtotal, coffeeUnitPrices, custState, cfg, in.LoyaltyPointsToUse, in.ManualDiscountPct)
	if in.Comp {
		disc = discountResult{Total: subtotal}
	}
	loyaltyPointsUsed := disc.PointsUsed
	total := disc.Total

	if !in.Comp {
		var paymentSum float64
		for _, p := range in.Payments {
			paymentSum += p.Amount
		}
		if math.Abs(paymentSum-total) > 0.01 {
			return nil, domain.NewBadRequest(fmt.Sprintf(
				"payment sum %.2f does not match order total %.2f", paymentSum, total))
		}
	}

	receiptNo := "000001"
	rawReceipt, err := q.GetNextReceiptNo(ctx, pgtype.UUID{Bytes: in.LocationID, Valid: true})
	if err == nil {
		if str, ok := rawReceipt.(string); ok && str != "" {
			receiptNo = str
		}
	}

	clientUUID := in.ClientUUID
	if clientUUID == (uuid.UUID{}) {
		clientUUID = uuid.New()
	}
	shiftID := pgtype.UUID{}
	if in.ShiftID != nil {
		shiftID = pgtype.UUID{Bytes: *in.ShiftID, Valid: true}
	}

	order, err := q.CreateOrder(ctx, pgdb.CreateOrderParams{
		LocationID:        pgtype.UUID{Bytes: in.LocationID, Valid: true},
		ShiftID:           shiftID,
		BaristaID:         pgtype.UUID{Bytes: in.BaristaID, Valid: true},
		CustomerID:        customerID,
		Subtotal:          numericFromFloat(subtotal),
		DiscountTotal:     numericFromFloat(disc.DiscountTotal),
		LoyaltyPointsUsed: numericFromFloat(loyaltyPointsUsed),
		Total:             numericFromFloat(total),
		CostTotal:         numericFromFloat(costTotal),
		ReceiptNo:         receiptNo,
		ClientUuid:        pgtype.UUID{Bytes: clientUUID, Valid: true},
		IsComp:            in.Comp,
		CompRecipient:     in.CompRecipient,
	})
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	orderID := pgtype.UUID{Bytes: uuid.UUID(order.ID.Bytes), Valid: true}

	var dtoItems []OrderItemDTO
	for _, ci := range calcItems {
		oi, err := q.CreateOrderItem(ctx, pgdb.CreateOrderItemParams{
			OrderID:           orderID,
			ProductID:         pgtype.UUID{Bytes: ci.productID, Valid: true},
			Qty:               numericFromFloat(ci.qty),
			UnitPriceSnapshot: numericFromFloat(ci.unitPrice),
			LineTotal:         numericFromFloat(ci.lineTotal),
			LineCost:          numericFromFloat(ci.lineCost),
			ClientUuid:        pgtype.UUID{Bytes: ci.clientUUID, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("create order item: %w", err)
		}
		itemID := pgtype.UUID{Bytes: uuid.UUID(oi.ID.Bytes), Valid: true}

		// Deduct ingredient stock for this line (recipe-driven consumption).
		for _, d := range ci.deductions {
			saleReason := "sale"
			if in.Comp {
				saleReason = "staff" // products taken off the shelf, not sold
			}
			if _, err := q.CreateStockMovement(ctx, pgdb.CreateStockMovementParams{
				LocationID:       pgtype.UUID{Bytes: in.LocationID, Valid: true},
				IngredientID:     pgtype.UUID{Bytes: d.ingredientID, Valid: true},
				QtyDelta:         numericFromFloat(d.qtyDelta),
				UnitCostSnapshot: numericFromFloat(d.costSnapshot),
				Reason:           saleReason,
				OrderID:          orderID,
				ClientUuid:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
				CreatedBy:        pgtype.UUID{Bytes: in.BaristaID, Valid: true},
			}); err != nil {
				return nil, fmt.Errorf("deduct stock: %w", err)
			}
		}

		var mods []OrderModifierDTO
		for j, modID := range ci.modIDs {
			oim, err := q.CreateOrderItemModifier(ctx, pgdb.CreateOrderItemModifierParams{
				OrderItemID:        itemID,
				ModifierOptionID:   pgtype.UUID{Bytes: modID, Valid: true},
				PriceDeltaSnapshot: numericFromFloat(ci.modDeltas[j]),
			})
			if err != nil {
				return nil, fmt.Errorf("create order item modifier: %w", err)
			}
			mods = append(mods, OrderModifierDTO{
				ID:               uuid.UUID(oim.ID.Bytes),
				ModifierOptionID: modID,
				OptionName:       ci.modNames[j],
				PriceDelta:       ci.modDeltas[j],
			})
		}
		dtoItems = append(dtoItems, OrderItemDTO{
			ID:                uuid.UUID(oi.ID.Bytes),
			ProductID:         ci.productID,
			ProductName:       ci.productName,
			Qty:               ci.qty,
			UnitPriceSnapshot: ci.unitPrice,
			LineTotal:         ci.lineTotal,
			ClientUUID:        ci.clientUUID,
			Modifiers:         mods,
		})
	}

	var dtoPays []PaymentDTO
	for _, p := range in.Payments {
		pm, ok := pmMap[p.MethodCode]
		if !ok {
			return nil, domain.NewBadRequest(fmt.Sprintf("unknown payment method: %s", p.MethodCode))
		}
		var extRef *string
		if p.ExternalRef != "" {
			extRef = &p.ExternalRef
		}
		pay, err := q.CreatePayment(ctx, pgdb.CreatePaymentParams{
			OrderID:         orderID,
			PaymentMethodID: pm.methodID,
			Amount:          numericFromFloat(p.Amount),
			ExternalRef:     extRef,
		})
		if err != nil {
			return nil, fmt.Errorf("create payment: %w", err)
		}
		dtoPays = append(dtoPays, PaymentDTO{
			ID:          uuid.UUID(pay.ID.Bytes),
			MethodCode:  p.MethodCode,
			MethodName:  pm.methodName,
			Amount:      p.Amount,
			ExternalRef: p.ExternalRef,
		})
	}

	if _, err := q.PayOrder(ctx, orderID); err != nil {
		return nil, fmt.Errorf("pay order: %w", err)
	}

	loyaltyEarned := 0.0
	if customerID.Valid {
		// New balance = prior - redeemed + earned on the paid amount.
		newBalance := custState.pointsBalance - disc.PointsUsed + disc.EarnedPoints
		if newBalance < 0 {
			newBalance = 0
		}
		if _, err := q.ApplyLoyaltyOrder(ctx, pgdb.ApplyLoyaltyOrderParams{
			CustomerID:     customerID,
			PointsBalance:  numericFromFloat(newBalance),
			FreeDrinksLeft: int32(disc.NewFreeDrinksLeft),
			CoffeePunches:  int32(disc.NewCoffeePunches),
		}); err != nil {
			return nil, fmt.Errorf("apply loyalty: %w", err)
		}
		loyaltyEarned = disc.EarnedPoints

		if loyaltyAccountID.Valid {
			recordLoyaltyTx(ctx, q, loyaltyAccountID, orderID, "earn", disc.EarnedPoints, receiptNo)
			recordLoyaltyTx(ctx, q, loyaltyAccountID, orderID, "redeem", -disc.PointsUsed, receiptNo)
			if disc.FreeCoffeesApplied > 0 {
				recordLoyaltyTx(ctx, q, loyaltyAccountID, orderID, "free_drink", 0, receiptNo)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	dto := orderToDTO(order)
	dto.Items = dtoItems
	dto.Payments = dtoPays
	dto.LoyaltyEarned = loyaltyEarned
	return &dto, nil
}

// CustomerProfile returns the current bonus balance and loyalty state for a phone number.
func (svc *OrderService) CustomerProfile(ctx context.Context, phone string) (*CustomerProfileDTO, error) {
	phone = normalizePhone(phone)
	if phone == "" {
		return &CustomerProfileDTO{}, nil
	}
	st, err := lookupLoyalty(ctx, svc.q, phone)
	if err != nil {
		return nil, fmt.Errorf("lookup loyalty: %w", err)
	}
	return &CustomerProfileDTO{
		Phone:          phone,
		CustomerFound:  st.found,
		Balance:        st.pointsBalance,
		FreeDrinksLeft: st.freeDrinksLeft,
		CoffeePunches:  st.coffeePunches,
	}, nil
}

// ─── Quote ────────────────────────────────────────────────────────────────────

type QuoteInput struct {
	CustomerPhone      string
	LoyaltyPointsToUse float64
	ManualDiscountPct  float64
	Items              []OrderItemInput
}

// QuoteDTO is the price breakdown shown in POS before payment is taken.
type QuoteDTO struct {
	Subtotal           float64 `json:"subtotal"`
	PromoDiscount      float64 `json:"promo_discount"`
	ManualDiscount     float64 `json:"manual_discount"`
	FreeCoffeeDiscount float64 `json:"free_coffee_discount"`
	FreeCoffeesApplied int     `json:"free_coffees_applied"`
	DiscountTotal      float64 `json:"discount_total"`
	LoyaltyPointsUsed  float64 `json:"loyalty_points_used"`
	Total              float64 `json:"total"`
	PromoActive        bool    `json:"promo_active"`
	PromoPercent       float64 `json:"promo_percent"`
	// Customer context (zero when no/new phone)
	CustomerFound  bool    `json:"customer_found"`
	PointsBalance  float64 `json:"points_balance"`
	FreeDrinksLeft int     `json:"free_drinks_left"`
}

// Quote prices a prospective order without persisting anything. It mirrors the
// pricing in CreateOrder so POS can collect the exact payable amount.
func (svc *OrderService) Quote(ctx context.Context, in QuoteInput) (*QuoteDTO, error) {
	cfg, err := loadLoyaltyConfig(ctx, svc.q)
	if err != nil {
		return nil, fmt.Errorf("load loyalty config: %w", err)
	}

	var subtotal float64
	var coffeeUnitPrices []float64
	for _, item := range in.Items {
		prod, err := svc.q.GetProduct(ctx, pgtype.UUID{Bytes: item.ProductID, Valid: true})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.NewNotFound("product")
			}
			return nil, fmt.Errorf("get product: %w", err)
		}
		unitPrice := floatFromNumeric(prod.BasePrice)
		if prod.SalePrice.Valid {
			unitPrice = floatFromNumeric(prod.SalePrice) // active sale event overrides base price
		}
		for _, modID := range item.ModifierOptionIDs {
			mod, err := svc.q.GetModifierOption(ctx, pgtype.UUID{Bytes: modID, Valid: true})
			if err != nil {
				return nil, fmt.Errorf("get modifier option: %w", err)
			}
			unitPrice += floatFromNumeric(mod.PriceDelta)
		}
		subtotal += math.Round(unitPrice*item.Qty*100) / 100
		if prod.CategoryName == cfg.FreeCategory {
			for n := 0; n < int(item.Qty); n++ {
				coffeeUnitPrices = append(coffeeUnitPrices, unitPrice)
			}
		}
	}

	var st loyaltyState
	if in.CustomerPhone != "" {
		st, err = lookupLoyalty(ctx, svc.q, in.CustomerPhone)
		if err != nil {
			return nil, fmt.Errorf("lookup loyalty: %w", err)
		}
	}

	disc := computeDiscounts(subtotal, coffeeUnitPrices, st, cfg, in.LoyaltyPointsToUse, in.ManualDiscountPct)
	return &QuoteDTO{
		Subtotal:           round2(subtotal),
		PromoDiscount:      disc.PromoDiscount,
		ManualDiscount:     disc.ManualDiscount,
		FreeCoffeeDiscount: disc.FreeCoffeeDiscount,
		FreeCoffeesApplied: disc.FreeCoffeesApplied,
		DiscountTotal:      disc.DiscountTotal,
		LoyaltyPointsUsed:  disc.PointsUsed,
		Total:              disc.Total,
		PromoActive:        cfg.PromoActive,
		PromoPercent:       cfg.PromoPct,
		CustomerFound:      st.found,
		PointsBalance:      st.pointsBalance,
		FreeDrinksLeft:     st.freeDrinksLeft,
	}, nil
}

// LoyaltyConfig returns the effective loyalty configuration.
func (svc *OrderService) LoyaltyConfig(ctx context.Context) (LoyaltyConfig, error) {
	return loadLoyaltyConfig(ctx, svc.q)
}

// SetPromo toggles/sets the global percent-off promo.
func (svc *OrderService) SetPromo(ctx context.Context, active bool, percent float64) (LoyaltyConfig, error) {
	params, _ := json.Marshal(map[string]float64{"percent": percent})
	if _, err := svc.q.SetLoyaltyRule(ctx, pgdb.SetLoyaltyRuleParams{
		Code:     rulePromoDiscount,
		Params:   params,
		IsActive: active,
	}); err != nil {
		return LoyaltyConfig{}, fmt.Errorf("set promo: %w", err)
	}
	return loadLoyaltyConfig(ctx, svc.q)
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (svc *OrderService) Get(ctx context.Context, id uuid.UUID) (*OrderDTO, error) {
	order, err := svc.q.GetOrder(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, domain.NewNotFound("order")
	}
	dto := orderToDTO(order)

	items, err := svc.q.GetOrderItems(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}

	allMods, err := svc.q.GetAllOrderItemModifiers(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("get order item modifiers: %w", err)
	}

	modsByItem := make(map[uuid.UUID][]OrderModifierDTO)
	for _, m := range allMods {
		itemID := uuid.UUID(m.OrderItemID.Bytes)
		modsByItem[itemID] = append(modsByItem[itemID], OrderModifierDTO{
			ID:               uuid.UUID(m.ID.Bytes),
			ModifierOptionID: uuid.UUID(m.ModifierOptionID.Bytes),
			OptionName:       m.OptionName,
			PriceDelta:       floatFromNumeric(m.PriceDeltaSnapshot),
		})
	}

	dtoItems := make([]OrderItemDTO, len(items))
	for i, item := range items {
		itemID := uuid.UUID(item.ID.Bytes)
		dtoItems[i] = OrderItemDTO{
			ID:                itemID,
			ProductID:         uuid.UUID(item.ProductID.Bytes),
			ProductName:       item.ProductName,
			Qty:               floatFromNumeric(item.Qty),
			UnitPriceSnapshot: floatFromNumeric(item.UnitPriceSnapshot),
			LineTotal:         floatFromNumeric(item.LineTotal),
			ClientUUID:        uuid.UUID(item.ClientUuid.Bytes),
			Modifiers:         modsByItem[itemID],
		}
	}
	dto.Items = dtoItems

	payments, err := svc.q.GetOrderPayments(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("get order payments: %w", err)
	}
	dtoPays := make([]PaymentDTO, len(payments))
	for i, p := range payments {
		extRef := ""
		if p.ExternalRef != nil {
			extRef = *p.ExternalRef
		}
		dtoPays[i] = PaymentDTO{
			ID:          uuid.UUID(p.ID.Bytes),
			MethodCode:  p.MethodCode,
			MethodName:  p.MethodName,
			Amount:      floatFromNumeric(p.Amount),
			ExternalRef: extRef,
		}
	}
	dto.Payments = dtoPays

	return &dto, nil
}

// ─── Receipt ──────────────────────────────────────────────────────────────────

func (svc *OrderService) Receipt(ctx context.Context, id uuid.UUID) (*ReceiptDTO, error) {
	order, err := svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	items := make([]ReceiptItemDTO, len(order.Items))
	for i, item := range order.Items {
		modNames := make([]string, len(item.Modifiers))
		for j, mod := range item.Modifiers {
			modNames[j] = mod.OptionName
		}
		items[i] = ReceiptItemDTO{
			Name:      item.ProductName,
			Qty:       item.Qty,
			UnitPrice: item.UnitPriceSnapshot,
			LineTotal: item.LineTotal,
			Modifiers: modNames,
		}
	}

	return &ReceiptDTO{
		ReceiptNo:     order.ReceiptNo,
		LocationID:    order.LocationID,
		BaristaID:     order.BaristaID,
		CreatedAt:     order.CreatedAt,
		Items:         items,
		Subtotal:      order.Subtotal,
		DiscountTotal: order.DiscountTotal,
		LoyaltyUsed:   order.LoyaltyPointsUsed,
		Total:         order.Total,
		Payments:      order.Payments,
		LoyaltyEarned: order.LoyaltyEarned,
	}, nil
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

func (svc *OrderService) Cancel(ctx context.Context, in CancelOrderInput) (*OrderDTO, error) {
	order, err := svc.q.CancelOrder(ctx, pgdb.CancelOrderParams{
		ID:           pgtype.UUID{Bytes: in.OrderID, Valid: true},
		CancelReason: &in.Reason,
		CancelledBy:  pgtype.UUID{Bytes: in.RequestedBy, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewBadRequest("order not found or cannot be cancelled")
		}
		return nil, fmt.Errorf("cancel order: %w", err)
	}
	dto := orderToDTO(order)
	return &dto, nil
}

// ─── Delete (admin) ─────────────────────────────────────────────────────────

// SoftDelete hides an order from listings without removing its data.
func (svc *OrderService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	oid := pgtype.UUID{Bytes: id, Valid: true}
	ct, err := svc.pool.Exec(ctx,
		`UPDATE orders SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, oid)
	if err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.NewNotFound("order")
	}
	return nil
}

// HardDelete permanently removes an order and reverses the stock consumption
// and loyalty changes it caused — for cleaning up test orders.
func (svc *OrderService) HardDelete(ctx context.Context, id uuid.UUID) error {
	tx, err := svc.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	oid := pgtype.UUID{Bytes: id, Valid: true}

	// 1. Restore ingredient stock: sale movements had negative qty_delta, so
	//    subtracting their sum adds the consumed quantity back.
	if _, err := tx.Exec(ctx, `
		UPDATE ingredients i
		SET current_qty = current_qty - sm.delta
		FROM (SELECT ingredient_id, COALESCE(SUM(qty_delta),0) AS delta
		      FROM stock_movements WHERE order_id = $1 GROUP BY ingredient_id) sm
		WHERE i.id = sm.ingredient_id`, oid); err != nil {
		return fmt.Errorf("restore stock: %w", err)
	}

	// 2. Reverse loyalty points and the visit count for this order.
	if _, err := tx.Exec(ctx, `
		UPDATE loyalty_accounts la
		SET points_balance = GREATEST(0, points_balance - lt.delta),
		    total_visits   = GREATEST(0, total_visits - 1),
		    updated_at     = NOW()
		FROM (SELECT loyalty_account_id, COALESCE(SUM(points_delta),0) AS delta
		      FROM loyalty_transactions WHERE order_id = $1 GROUP BY loyalty_account_id) lt
		WHERE la.id = lt.loyalty_account_id`, oid); err != nil {
		return fmt.Errorf("reverse loyalty: %w", err)
	}

	// 3. Remove non-cascading children, then the order (cascades items/modifiers/payments).
	for _, stmt := range []string{
		`DELETE FROM loyalty_transactions WHERE order_id = $1`,
		`DELETE FROM stock_movements WHERE order_id = $1`,
		`DELETE FROM refunds WHERE order_id = $1`,
	} {
		if _, err := tx.Exec(ctx, stmt, oid); err != nil {
			return fmt.Errorf("delete children: %w", err)
		}
	}
	ct, err := tx.Exec(ctx, `DELETE FROM orders WHERE id = $1`, oid)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.NewNotFound("order")
	}
	return tx.Commit(ctx)
}

// ─── List ─────────────────────────────────────────────────────────────────────

type ListOrderDTO struct {
	ID                uuid.UUID  `json:"id"`
	LocationID        uuid.UUID  `json:"location_id"`
	ShiftID           *uuid.UUID `json:"shift_id,omitempty"`
	BaristaID         uuid.UUID  `json:"barista_id"`
	BaristaName       string     `json:"barista_name"`
	Status            string     `json:"status"`
	Subtotal          float64    `json:"subtotal"`
	LoyaltyPointsUsed float64    `json:"loyalty_points_used"`
	Total             float64    `json:"total"`
	ReceiptNo         string     `json:"receipt_no"`
	CreatedAt         time.Time  `json:"created_at"`
}

func (svc *OrderService) ListByShift(ctx context.Context, shiftID uuid.UUID) ([]ListOrderDTO, error) {
	rows, err := svc.q.ListOrdersByShift(ctx, pgtype.UUID{Bytes: shiftID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list orders by shift: %w", err)
	}
	out := make([]ListOrderDTO, len(rows))
	for i, r := range rows {
		out[i] = listOrderFromShiftRow(r)
	}
	return out, nil
}

func (svc *OrderService) ListByLocation(ctx context.Context, locationID uuid.UUID, from, to time.Time) ([]ListOrderDTO, error) {
	rows, err := svc.q.ListOrdersByLocation(ctx, pgdb.ListOrdersByLocationParams{
		LocationID:  pgtype.UUID{Bytes: locationID, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: from, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: to, Valid: true},
		Limit:       200,
		Offset:      0,
	})
	if err != nil {
		return nil, fmt.Errorf("list orders by location: %w", err)
	}
	out := make([]ListOrderDTO, len(rows))
	for i, r := range rows {
		out[i] = listOrderFromLocationRow(r)
	}
	return out, nil
}

func listOrderFromShiftRow(r pgdb.ListOrdersByShiftRow) ListOrderDTO {
	dto := ListOrderDTO{
		ID:                uuid.UUID(r.ID.Bytes),
		LocationID:        uuid.UUID(r.LocationID.Bytes),
		BaristaID:         uuid.UUID(r.BaristaID.Bytes),
		BaristaName:       r.BaristaName,
		Status:            r.Status,
		Subtotal:          floatFromNumeric(r.Subtotal),
		LoyaltyPointsUsed: floatFromNumeric(r.LoyaltyPointsUsed),
		Total:             floatFromNumeric(r.Total),
		ReceiptNo:         r.ReceiptNo,
		CreatedAt:         r.CreatedAt.Time,
	}
	if r.ShiftID.Valid {
		id := uuid.UUID(r.ShiftID.Bytes)
		dto.ShiftID = &id
	}
	return dto
}

func listOrderFromLocationRow(r pgdb.ListOrdersByLocationRow) ListOrderDTO {
	dto := ListOrderDTO{
		ID:                uuid.UUID(r.ID.Bytes),
		LocationID:        uuid.UUID(r.LocationID.Bytes),
		BaristaID:         uuid.UUID(r.BaristaID.Bytes),
		BaristaName:       r.BaristaName,
		Status:            r.Status,
		Subtotal:          floatFromNumeric(r.Subtotal),
		LoyaltyPointsUsed: floatFromNumeric(r.LoyaltyPointsUsed),
		Total:             floatFromNumeric(r.Total),
		ReceiptNo:         r.ReceiptNo,
		CreatedAt:         r.CreatedAt.Time,
	}
	if r.ShiftID.Valid {
		id := uuid.UUID(r.ShiftID.Bytes)
		dto.ShiftID = &id
	}
	return dto
}

// ─── ListPaymentMethods ───────────────────────────────────────────────────────

type PaymentMethodDTO struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	OpensDrawer bool      `json:"opens_drawer"`
}

func (svc *OrderService) ListPaymentMethods(ctx context.Context) ([]PaymentMethodDTO, error) {
	methods, err := svc.q.ListPaymentMethods(ctx)
	if err != nil {
		return nil, fmt.Errorf("list payment methods: %w", err)
	}
	out := make([]PaymentMethodDTO, len(methods))
	for i, m := range methods {
		out[i] = PaymentMethodDTO{
			ID:          uuid.UUID(m.ID.Bytes),
			Code:        m.Code,
			Name:        m.Name,
			OpensDrawer: m.OpensDrawer,
		}
	}
	return out, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

type pmInfo struct {
	methodID   pgtype.UUID
	methodName string
}

func loadPaymentMethods(ctx context.Context, q *pgdb.Queries) (map[string]pmInfo, error) {
	methods, err := q.ListPaymentMethods(ctx)
	if err != nil {
		return nil, fmt.Errorf("load payment methods: %w", err)
	}
	result := make(map[string]pmInfo, len(methods))
	for _, m := range methods {
		result[m.Code] = pmInfo{methodID: m.ID, methodName: m.Name}
	}
	return result, nil
}

// getOrCreateCustomer ensures a customer + loyalty account exist for the phone
// and returns the customer id, loyalty account id, and the account snapshot.
func getOrCreateCustomer(ctx context.Context, q *pgdb.Queries, phone string) (uuid.UUID, uuid.UUID, loyaltyState, error) {
	phone = normalizePhone(phone)
	cust, err := q.GetCustomerByPhone(ctx, phone)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, uuid.UUID{}, loyaltyState{}, fmt.Errorf("get customer: %w", err)
		}
		cust, err = q.CreateCustomer(ctx, pgdb.CreateCustomerParams{Phone: phone, Name: ""})
		if err != nil {
			return uuid.UUID{}, uuid.UUID{}, loyaltyState{}, fmt.Errorf("create customer: %w", err)
		}
	}

	custUID := uuid.UUID(cust.ID.Bytes)
	loyAcct, err := q.GetLoyaltyAccountByCustomer(ctx, pgtype.UUID{Bytes: custUID, Valid: true})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, uuid.UUID{}, loyaltyState{}, fmt.Errorf("get loyalty account: %w", err)
		}
		loyAcct, err = q.CreateLoyaltyAccount(ctx, pgtype.UUID{Bytes: custUID, Valid: true})
		if err != nil {
			return uuid.UUID{}, uuid.UUID{}, loyaltyState{}, fmt.Errorf("create loyalty account: %w", err)
		}
	}

	st := loyaltyState{
		found:          true,
		customerID:     cust.ID.Bytes,
		pointsBalance:  floatFromNumeric(loyAcct.PointsBalance),
		freeDrinksLeft: int(loyAcct.FreeDrinksLeft),
		coffeePunches:  int(loyAcct.CoffeePunches),
	}
	return custUID, uuid.UUID(loyAcct.ID.Bytes), st, nil
}

// recordLoyaltyTx best-effort logs a loyalty ledger entry (non-fatal).
func recordLoyaltyTx(ctx context.Context, q *pgdb.Queries, acctID, orderID pgtype.UUID, kind string, pointsDelta float64, receiptNo string) {
	if kind == "earn" && pointsDelta == 0 {
		return
	}
	if kind == "redeem" && pointsDelta == 0 {
		return
	}
	note := fmt.Sprintf("Чек %s", receiptNo)
	_, _ = q.CreateLoyaltyTransaction(ctx, pgdb.CreateLoyaltyTransactionParams{
		LoyaltyAccountID: acctID,
		OrderID:          orderID,
		Kind:             kind,
		PointsDelta:      numericFromFloat(pointsDelta),
		Note:             &note,
	})
}

func orderToDTO(o pgdb.Order) OrderDTO {
	dto := OrderDTO{
		ID:                uuid.UUID(o.ID.Bytes),
		LocationID:        uuid.UUID(o.LocationID.Bytes),
		BaristaID:         uuid.UUID(o.BaristaID.Bytes),
		Status:            o.Status,
		Subtotal:          floatFromNumeric(o.Subtotal),
		DiscountTotal:     floatFromNumeric(o.DiscountTotal),
		LoyaltyPointsUsed: floatFromNumeric(o.LoyaltyPointsUsed),
		Total:             floatFromNumeric(o.Total),
		ReceiptNo:         o.ReceiptNo,
		ClientUUID:        uuid.UUID(o.ClientUuid.Bytes),
		CreatedAt:         o.CreatedAt.Time,
	}
	if o.ShiftID.Valid {
		id := uuid.UUID(o.ShiftID.Bytes)
		dto.ShiftID = &id
	}
	if o.CustomerID.Valid {
		id := uuid.UUID(o.CustomerID.Bytes)
		dto.CustomerID = &id
	}
	return dto
}
