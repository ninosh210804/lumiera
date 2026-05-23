package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pgdb "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/domain"
)

// RecipeService manages tech cards (recipes) and product linking.
type RecipeService struct {
	q *pgdb.Queries
}

func NewRecipeService(q *pgdb.Queries) *RecipeService {
	return &RecipeService{q: q}
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type RecipeDTO struct {
	ID         uuid.UUID       `json:"id"`
	LocationID uuid.UUID       `json:"location_id"`
	Name       string          `json:"name"`
	RecipeType string          `json:"recipe_type"`
	YieldQty   float64         `json:"yield_qty"`
	YieldUnit  string          `json:"yield_unit"`
	Notes      string          `json:"notes"`
	Items      []RecipeItemDTO `json:"items,omitempty"`
	TotalCost  float64         `json:"total_cost,omitempty"`
}

type RecipeItemDTO struct {
	ID             uuid.UUID  `json:"id"`
	RecipeID       uuid.UUID  `json:"recipe_id"`
	IngredientID   *uuid.UUID `json:"ingredient_id,omitempty"`
	SubRecipeID    *uuid.UUID `json:"sub_recipe_id,omitempty"`
	IngredientName *string    `json:"ingredient_name,omitempty"`
	SubRecipeName  *string    `json:"sub_recipe_name,omitempty"`
	Qty            float64    `json:"qty"`
	Unit           string     `json:"unit"`
	SortOrder      int32      `json:"sort_order"`
	UnitCost       float64    `json:"unit_cost"`
	LineCost       float64    `json:"line_cost"`
}

// ─── Queries ──────────────────────────────────────────────────────────────────

func (s *RecipeService) List(ctx context.Context, locationID uuid.UUID) ([]RecipeDTO, error) {
	rows, err := s.q.ListRecipes(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list recipes: %w", err)
	}
	out := make([]RecipeDTO, len(rows))
	for i, r := range rows {
		out[i] = recipeToDTO(r)
	}
	return out, nil
}

func (s *RecipeService) Get(ctx context.Context, id uuid.UUID) (*RecipeDTO, error) {
	r, err := s.q.GetRecipe(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, domain.NewNotFound("recipe")
	}
	dto := recipeToDTO(r)

	items, err := s.q.GetRecipeItems(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("get recipe items: %w", err)
	}
	dto.Items = make([]RecipeItemDTO, len(items))
	var totalCost float64
	for i, item := range items {
		ri := recipeItemToDTO(item)
		dto.Items[i] = ri
		totalCost += ri.LineCost
	}
	dto.TotalCost = totalCost
	return &dto, nil
}

// ─── Recipe mutations ─────────────────────────────────────────────────────────

type CreateRecipeInput struct {
	LocationID uuid.UUID
	Name       string
	RecipeType string
	YieldQty   float64
	YieldUnit  string
	Notes      string
	CreatedBy  uuid.UUID
}

func (s *RecipeService) Create(ctx context.Context, in CreateRecipeInput) (*RecipeDTO, error) {
	if in.RecipeType == "" {
		in.RecipeType = "product"
	}
	if in.YieldUnit == "" {
		in.YieldUnit = "pcs"
	}
	if in.YieldQty <= 0 {
		in.YieldQty = 1
	}
	r, err := s.q.CreateRecipe(ctx, pgdb.CreateRecipeParams{
		LocationID: pgtype.UUID{Bytes: in.LocationID, Valid: true},
		Name:       in.Name,
		RecipeType: in.RecipeType,
		YieldQty:   numericFromFloat(in.YieldQty),
		YieldUnit:  in.YieldUnit,
		Notes:      in.Notes,
		CreatedBy:  pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create recipe: %w", err)
	}
	dto := recipeToDTO(r)
	dto.Items = []RecipeItemDTO{}
	return &dto, nil
}

type UpdateRecipeInput struct {
	ID         uuid.UUID
	Name       string
	RecipeType string
	YieldQty   float64
	YieldUnit  string
	Notes      string
}

func (s *RecipeService) Update(ctx context.Context, in UpdateRecipeInput) (*RecipeDTO, error) {
	r, err := s.q.UpdateRecipe(ctx, pgdb.UpdateRecipeParams{
		ID:         pgtype.UUID{Bytes: in.ID, Valid: true},
		Name:       in.Name,
		RecipeType: in.RecipeType,
		YieldQty:   numericFromFloat(in.YieldQty),
		YieldUnit:  in.YieldUnit,
		Notes:      in.Notes,
	})
	if err != nil {
		return nil, fmt.Errorf("update recipe: %w", err)
	}
	dto := recipeToDTO(r)
	return &dto, nil
}

func (s *RecipeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.q.SoftDeleteRecipe(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

// LinkToProduct sets products.recipe_id for a product.
func (s *RecipeService) LinkToProduct(ctx context.Context, productID, recipeID uuid.UUID) error {
	return s.q.LinkProductRecipe(ctx, pgdb.LinkProductRecipeParams{
		ID:       pgtype.UUID{Bytes: productID, Valid: true},
		RecipeID: pgtype.UUID{Bytes: recipeID, Valid: true},
	})
}

// UnlinkFromProduct clears products.recipe_id.
func (s *RecipeService) UnlinkFromProduct(ctx context.Context, productID uuid.UUID) error {
	return s.q.UnlinkProductRecipe(ctx, pgtype.UUID{Bytes: productID, Valid: true})
}

// ─── Recipe item mutations ────────────────────────────────────────────────────

type AddRecipeItemInput struct {
	RecipeID     uuid.UUID
	IngredientID *uuid.UUID
	SubRecipeID  *uuid.UUID
	Qty          float64
	Unit         string
	SortOrder    int32
}

func (s *RecipeService) AddItem(ctx context.Context, in AddRecipeItemInput) (*RecipeItemDTO, error) {
	if in.IngredientID == nil && in.SubRecipeID == nil {
		return nil, domain.NewBadRequest("one of ingredient_id or sub_recipe_id is required")
	}
	if in.IngredientID != nil && in.SubRecipeID != nil {
		return nil, domain.NewBadRequest("only one of ingredient_id or sub_recipe_id may be set")
	}

	ingredientID := pgtype.UUID{}
	if in.IngredientID != nil {
		ingredientID = pgtype.UUID{Bytes: *in.IngredientID, Valid: true}
	}
	subRecipeID := pgtype.UUID{}
	if in.SubRecipeID != nil {
		subRecipeID = pgtype.UUID{Bytes: *in.SubRecipeID, Valid: true}
	}

	item, err := s.q.AddRecipeItem(ctx, pgdb.AddRecipeItemParams{
		RecipeID:     pgtype.UUID{Bytes: in.RecipeID, Valid: true},
		IngredientID: ingredientID,
		SubRecipeID:  subRecipeID,
		Qty:          numericFromFloat(in.Qty),
		Unit:         in.Unit,
		SortOrder:    in.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("add recipe item: %w", err)
	}
	dto := recipeItemFromRaw(item)
	return &dto, nil
}

type UpdateRecipeItemInput struct {
	ID        uuid.UUID
	Qty       float64
	Unit      string
	SortOrder int32
}

func (s *RecipeService) UpdateItem(ctx context.Context, in UpdateRecipeItemInput) (*RecipeItemDTO, error) {
	item, err := s.q.UpdateRecipeItem(ctx, pgdb.UpdateRecipeItemParams{
		ID:        pgtype.UUID{Bytes: in.ID, Valid: true},
		Qty:       numericFromFloat(in.Qty),
		Unit:      in.Unit,
		SortOrder: in.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("update recipe item: %w", err)
	}
	dto := recipeItemFromRaw(item)
	return &dto, nil
}

func (s *RecipeService) RemoveItem(ctx context.Context, itemID uuid.UUID) error {
	return s.q.RemoveRecipeItem(ctx, pgtype.UUID{Bytes: itemID, Valid: true})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func recipeToDTO(r pgdb.Recipe) RecipeDTO {
	return RecipeDTO{
		ID:         uuid.UUID(r.ID.Bytes),
		LocationID: uuid.UUID(r.LocationID.Bytes),
		Name:       r.Name,
		RecipeType: r.RecipeType,
		YieldQty:   floatFromNumeric(r.YieldQty),
		YieldUnit:  r.YieldUnit,
		Notes:      r.Notes,
	}
}

func recipeItemToDTO(r pgdb.GetRecipeItemsRow) RecipeItemDTO {
	dto := RecipeItemDTO{
		ID:             uuid.UUID(r.ID.Bytes),
		RecipeID:       uuid.UUID(r.RecipeID.Bytes),
		Qty:            floatFromNumeric(r.Qty),
		Unit:           r.Unit,
		SortOrder:      r.SortOrder,
		IngredientName: r.IngredientName,
		SubRecipeName:  r.SubRecipeName,
		UnitCost:       floatFromNumeric(r.CurrentAvgCost),
	}
	if r.IngredientID.Valid {
		id := uuid.UUID(r.IngredientID.Bytes)
		dto.IngredientID = &id
	}
	if r.SubRecipeID.Valid {
		id := uuid.UUID(r.SubRecipeID.Bytes)
		dto.SubRecipeID = &id
	}
	dto.LineCost = dto.Qty * dto.UnitCost
	return dto
}

func recipeItemFromRaw(r pgdb.RecipeItem) RecipeItemDTO {
	dto := RecipeItemDTO{
		ID:        uuid.UUID(r.ID.Bytes),
		RecipeID:  uuid.UUID(r.RecipeID.Bytes),
		Qty:       floatFromNumeric(r.Qty),
		Unit:      r.Unit,
		SortOrder: r.SortOrder,
	}
	if r.IngredientID.Valid {
		id := uuid.UUID(r.IngredientID.Bytes)
		dto.IngredientID = &id
	}
	if r.SubRecipeID.Valid {
		id := uuid.UUID(r.SubRecipeID.Bytes)
		dto.SubRecipeID = &id
	}
	return dto
}
