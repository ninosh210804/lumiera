package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

// MenuService handles categories, products, and modifiers.
type MenuService struct {
	q *pgdb.Queries
}

func NewMenuService(q *pgdb.Queries) *MenuService {
	return &MenuService{q: q}
}

// ─── Categories ───────────────────────────────────────────────────────────────

type CategoryDTO struct {
	ID         uuid.UUID `json:"id"`
	LocationID uuid.UUID `json:"location_id"`
	Name       string    `json:"name"`
	Icon       string    `json:"icon"`
	SortOrder  int32     `json:"sort_order"`
	IsActive   bool      `json:"is_active"`
}

func (s *MenuService) ListCategories(ctx context.Context, locationID uuid.UUID) ([]CategoryDTO, error) {
	rows, err := s.q.ListCategories(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	out := make([]CategoryDTO, len(rows))
	for i, r := range rows {
		out[i] = categoryToDTO(r)
	}
	return out, nil
}

type CreateCategoryInput struct {
	LocationID uuid.UUID
	Name       string
	Icon       string
	SortOrder  int32
	CreatedBy  uuid.UUID
}

func (s *MenuService) CreateCategory(ctx context.Context, in CreateCategoryInput) (*CategoryDTO, error) {
	cat, err := s.q.CreateCategory(ctx, pgdb.CreateCategoryParams{
		LocationID: pgtype.UUID{Bytes: in.LocationID, Valid: true},
		Name:       in.Name,
		Icon:       in.Icon,
		SortOrder:  in.SortOrder,
		CreatedBy:  pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	dto := categoryToDTO(cat)
	return &dto, nil
}

type UpdateCategoryInput struct {
	ID        uuid.UUID
	Name      string
	Icon      string
	SortOrder int32
	IsActive  bool
}

func (s *MenuService) UpdateCategory(ctx context.Context, in UpdateCategoryInput) (*CategoryDTO, error) {
	cat, err := s.q.UpdateCategory(ctx, pgdb.UpdateCategoryParams{
		ID:        pgtype.UUID{Bytes: in.ID, Valid: true},
		Name:      in.Name,
		Icon:      in.Icon,
		SortOrder: in.SortOrder,
		IsActive:  in.IsActive,
	})
	if err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	dto := categoryToDTO(cat)
	return &dto, nil
}

func (s *MenuService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.q.SoftDeleteCategory(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

// ─── Products ─────────────────────────────────────────────────────────────────

type ProductDTO struct {
	ID           uuid.UUID `json:"id"`
	LocationID   uuid.UUID `json:"location_id"`
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	SKU          *string   `json:"sku"`
	BasePrice    float64   `json:"base_price"`
	IsActive     bool      `json:"is_active"`
	IsStopListed bool      `json:"is_stop_listed"`
	ImageURL     *string    `json:"image_url"`
	SortOrder    int32      `json:"sort_order"`
	RecipeID     *uuid.UUID `json:"recipe_id"`
	SalePrice    *float64   `json:"sale_price"`
}

type ProductDetailDTO struct {
	ProductDTO
	ModifierGroups []ModifierGroupDTO `json:"modifier_groups"`
}

func (s *MenuService) ListProducts(ctx context.Context, locationID uuid.UUID, activeOnly bool) ([]ProductDTO, error) {
	pgLoc := pgtype.UUID{Bytes: locationID, Valid: true}
	if activeOnly {
		rows, err := s.q.ListActiveProducts(ctx, pgLoc)
		if err != nil {
			return nil, fmt.Errorf("list active products: %w", err)
		}
		out := make([]ProductDTO, len(rows))
		for i, r := range rows {
			out[i] = productRowToDTO(r.ID, r.LocationID, r.CategoryID, r.Name, r.Description,
				r.Sku, r.BasePrice, r.IsActive, r.IsStopListed, r.ImageUrl, r.SortOrder, r.CategoryName, r.RecipeID, r.SalePrice)
		}
		return out, nil
	}
	rows, err := s.q.ListProducts(ctx, pgLoc)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	out := make([]ProductDTO, len(rows))
	for i, r := range rows {
		out[i] = productRowToDTO(r.ID, r.LocationID, r.CategoryID, r.Name, r.Description,
			r.Sku, r.BasePrice, r.IsActive, r.IsStopListed, r.ImageUrl, r.SortOrder, r.CategoryName, r.RecipeID, r.SalePrice)
	}
	return out, nil
}

// ListMenu returns the full active menu (for POS) grouped by category.
func (s *MenuService) ListMenu(ctx context.Context, locationID uuid.UUID) ([]ProductDTO, error) {
	rows, err := s.q.ListActiveMenuProducts(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list menu: %w", err)
	}
	out := make([]ProductDTO, len(rows))
	for i, r := range rows {
		out[i] = productRowToDTO(r.ID, r.LocationID, r.CategoryID, r.Name, r.Description,
			r.Sku, r.BasePrice, r.IsActive, r.IsStopListed, r.ImageUrl, r.SortOrder, r.CategoryName, r.RecipeID, r.SalePrice)
	}
	return out, nil
}

func (s *MenuService) GetProduct(ctx context.Context, id uuid.UUID) (*ProductDetailDTO, error) {
	row, err := s.q.GetProduct(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, domain.NewNotFound("product")
	}
	dto := &ProductDetailDTO{
		ProductDTO: productRowToDTO(row.ID, row.LocationID, row.CategoryID, row.Name, row.Description,
			row.Sku, row.BasePrice, row.IsActive, row.IsStopListed, row.ImageUrl, row.SortOrder, row.CategoryName, row.RecipeID, row.SalePrice),
	}

	groups, err := s.q.GetModifierGroups(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("get modifier groups: %w", err)
	}
	dto.ModifierGroups = make([]ModifierGroupDTO, 0, len(groups))
	for _, g := range groups {
		mgDTO := ModifierGroupDTO{
			ID:            uuid.UUID(g.ID.Bytes),
			ProductID:     uuid.UUID(g.ProductID.Bytes),
			Name:          g.Name,
			SelectionType: g.SelectionType,
			Required:      g.Required,
			MinSelect:     g.MinSelect,
			MaxSelect:     g.MaxSelect,
			SortOrder:     g.SortOrder,
		}
		_ = json.Unmarshal(g.Options, &mgDTO.Options)
		dto.ModifierGroups = append(dto.ModifierGroups, mgDTO)
	}
	return dto, nil
}

type CreateProductInput struct {
	LocationID  uuid.UUID
	CategoryID  uuid.UUID
	Name        string
	Description string
	SKU         *string
	BasePrice   float64
	IsActive    bool
	ImageURL    *string
	SortOrder   int32
	CreatedBy   uuid.UUID
}

func (s *MenuService) CreateProduct(ctx context.Context, in CreateProductInput) (*ProductDTO, error) {
	p, err := s.q.CreateProduct(ctx, pgdb.CreateProductParams{
		LocationID:  pgtype.UUID{Bytes: in.LocationID, Valid: true},
		CategoryID:  pgtype.UUID{Bytes: in.CategoryID, Valid: true},
		Name:        in.Name,
		Description: in.Description,
		Sku:         in.SKU,
		BasePrice:   numericFromFloat(in.BasePrice),
		IsActive:    in.IsActive,
		ImageUrl:    in.ImageURL,
		SortOrder:   in.SortOrder,
		CreatedBy:   pgtype.UUID{Bytes: in.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}
	dto := productToDTO(p)
	return &dto, nil
}

type UpdateProductInput struct {
	ID          uuid.UUID
	Name        string
	Description string
	CategoryID  uuid.UUID
	BasePrice   float64
	IsActive    bool
	ImageURL    *string
	SortOrder   int32
	SKU         *string
}

func (s *MenuService) UpdateProduct(ctx context.Context, in UpdateProductInput) (*ProductDTO, error) {
	p, err := s.q.UpdateProduct(ctx, pgdb.UpdateProductParams{
		ID:          pgtype.UUID{Bytes: in.ID, Valid: true},
		Name:        in.Name,
		Description: in.Description,
		CategoryID:  pgtype.UUID{Bytes: in.CategoryID, Valid: true},
		BasePrice:   numericFromFloat(in.BasePrice),
		IsActive:    in.IsActive,
		ImageUrl:    in.ImageURL,
		SortOrder:   in.SortOrder,
		Sku:         in.SKU,
	})
	if err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}
	dto := productToDTO(p)
	return &dto, nil
}

func (s *MenuService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.q.SoftDeleteProduct(ctx, pgtype.UUID{Bytes: id, Valid: true})
}

func (s *MenuService) SetStopList(ctx context.Context, id uuid.UUID, stopped bool) error {
	return s.q.SetProductStopList(ctx, pgdb.SetProductStopListParams{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		IsStopListed: stopped,
	})
}

// ─── Modifiers ────────────────────────────────────────────────────────────────

type ModifierGroupDTO struct {
	ID            uuid.UUID           `json:"id"`
	ProductID     uuid.UUID           `json:"product_id"`
	Name          string              `json:"name"`
	SelectionType string              `json:"selection_type"`
	Required      bool                `json:"required"`
	MinSelect     int32               `json:"min_select"`
	MaxSelect     int32               `json:"max_select"`
	SortOrder     int32               `json:"sort_order"`
	Options       []ModifierOptionDTO `json:"options"`
}

type ModifierOptionDTO struct {
	ID                 uuid.UUID  `json:"id"`
	Name               string     `json:"name"`
	PriceDelta         float64    `json:"price_delta"`
	IsActive           bool       `json:"is_active"`
	SortOrder          int32      `json:"sort_order"`
	LinkedIngredientID *uuid.UUID `json:"linked_ingredient_id,omitempty"`
	IngredientQtyDelta float64    `json:"ingredient_qty_delta,omitempty"`
}

type CreateModifierGroupInput struct {
	ProductID     uuid.UUID
	Name          string
	SelectionType string
	Required      bool
	MinSelect     int32
	MaxSelect     int32
	SortOrder     int32
}

func (s *MenuService) CreateModifierGroup(ctx context.Context, in CreateModifierGroupInput) (*ModifierGroupDTO, error) {
	g, err := s.q.CreateModifierGroup(ctx, pgdb.CreateModifierGroupParams{
		ProductID:     pgtype.UUID{Bytes: in.ProductID, Valid: true},
		Name:          in.Name,
		SelectionType: in.SelectionType,
		Required:      in.Required,
		MinSelect:     in.MinSelect,
		MaxSelect:     in.MaxSelect,
		SortOrder:     in.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("create modifier group: %w", err)
	}
	return &ModifierGroupDTO{
		ID:            uuid.UUID(g.ID.Bytes),
		ProductID:     uuid.UUID(g.ProductID.Bytes),
		Name:          g.Name,
		SelectionType: g.SelectionType,
		Required:      g.Required,
		MinSelect:     g.MinSelect,
		MaxSelect:     g.MaxSelect,
		SortOrder:     g.SortOrder,
		Options:       []ModifierOptionDTO{},
	}, nil
}

func (s *MenuService) DeleteModifierGroup(ctx context.Context, groupID uuid.UUID) error {
	return s.q.DeleteModifierGroup(ctx, pgtype.UUID{Bytes: groupID, Valid: true})
}

type CreateModifierOptionInput struct {
	GroupID            uuid.UUID
	Name               string
	PriceDelta         float64
	LinkedIngredientID *uuid.UUID
	IngredientQtyDelta float64
	SortOrder          int32
}

func (s *MenuService) CreateModifierOption(ctx context.Context, in CreateModifierOptionInput) (*ModifierOptionDTO, error) {
	linkedID := pgtype.UUID{}
	if in.LinkedIngredientID != nil {
		linkedID = pgtype.UUID{Bytes: *in.LinkedIngredientID, Valid: true}
	}
	opt, err := s.q.CreateModifierOption(ctx, pgdb.CreateModifierOptionParams{
		GroupID:            pgtype.UUID{Bytes: in.GroupID, Valid: true},
		Name:               in.Name,
		PriceDelta:         numericFromFloat(in.PriceDelta),
		LinkedIngredientID: linkedID,
		IngredientQtyDelta: numericFromFloat(in.IngredientQtyDelta),
		SortOrder:          in.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("create modifier option: %w", err)
	}
	return modifierOptionToDTO(opt), nil
}

type UpdateModifierOptionInput struct {
	ID         uuid.UUID
	Name       string
	PriceDelta float64
	IsActive   bool
	SortOrder  int32
}

func (s *MenuService) UpdateModifierOption(ctx context.Context, in UpdateModifierOptionInput) (*ModifierOptionDTO, error) {
	opt, err := s.q.UpdateModifierOption(ctx, pgdb.UpdateModifierOptionParams{
		ID:         pgtype.UUID{Bytes: in.ID, Valid: true},
		Name:       in.Name,
		PriceDelta: numericFromFloat(in.PriceDelta),
		IsActive:   in.IsActive,
		SortOrder:  in.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("update modifier option: %w", err)
	}
	return modifierOptionToDTO(opt), nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func numericFromFloat(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(fmt.Sprintf("%.4f", f))
	return n
}

func floatFromNumeric(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	v, _ := n.Float64Value()
	return v.Float64
}

func categoryToDTO(c pgdb.Category) CategoryDTO {
	return CategoryDTO{
		ID:         uuid.UUID(c.ID.Bytes),
		LocationID: uuid.UUID(c.LocationID.Bytes),
		Name:       c.Name,
		Icon:       c.Icon,
		SortOrder:  c.SortOrder,
		IsActive:   c.IsActive,
	}
}

func productRowToDTO(
	id, locationID, categoryID pgtype.UUID,
	name, description string,
	sku *string,
	basePrice pgtype.Numeric,
	isActive, isStopListed bool,
	imageURL *string,
	sortOrder int32,
	categoryName string,
	recipeID pgtype.UUID,
	salePrice pgtype.Numeric,
) ProductDTO {
	dto := ProductDTO{
		ID:           uuid.UUID(id.Bytes),
		LocationID:   uuid.UUID(locationID.Bytes),
		CategoryID:   uuid.UUID(categoryID.Bytes),
		CategoryName: categoryName,
		Name:         name,
		Description:  description,
		SKU:          sku,
		BasePrice:    floatFromNumeric(basePrice),
		IsActive:     isActive,
		IsStopListed: isStopListed,
		ImageURL:     imageURL,
		SortOrder:    sortOrder,
	}
	if recipeID.Valid {
		rid := uuid.UUID(recipeID.Bytes)
		dto.RecipeID = &rid
	}
	if salePrice.Valid {
		sp := floatFromNumeric(salePrice)
		dto.SalePrice = &sp
	}
	return dto
}

func productToDTO(p pgdb.Product) ProductDTO {
	return ProductDTO{
		ID:           uuid.UUID(p.ID.Bytes),
		LocationID:   uuid.UUID(p.LocationID.Bytes),
		CategoryID:   uuid.UUID(p.CategoryID.Bytes),
		Name:         p.Name,
		Description:  p.Description,
		SKU:          p.Sku,
		BasePrice:    floatFromNumeric(p.BasePrice),
		IsActive:     p.IsActive,
		IsStopListed: p.IsStopListed,
		ImageURL:     p.ImageUrl,
		SortOrder:    p.SortOrder,
	}
}

func modifierOptionToDTO(opt pgdb.ModifierOption) *ModifierOptionDTO {
	dto := &ModifierOptionDTO{
		ID:                 uuid.UUID(opt.ID.Bytes),
		Name:               opt.Name,
		PriceDelta:         floatFromNumeric(opt.PriceDelta),
		IsActive:           opt.IsActive,
		SortOrder:          opt.SortOrder,
		IngredientQtyDelta: floatFromNumeric(opt.IngredientQtyDelta),
	}
	if opt.LinkedIngredientID.Valid {
		id := uuid.UUID(opt.LinkedIngredientID.Bytes)
		dto.LinkedIngredientID = &id
	}
	return dto
}
