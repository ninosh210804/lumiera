package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	mw "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/middleware"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/service"
)

type menuHandler struct {
	menu *service.MenuService
}

func locationFromCtxOrQuery(r *http.Request) (uuid.UUID, error) {
	if locStr := r.URL.Query().Get("location_id"); locStr != "" {
		return uuid.Parse(locStr)
	}
	if claims, ok := mw.ClaimsFrom(r.Context()); ok {
		return claims.LocationID, nil
	}
	return uuid.UUID{}, badRequestf("location_id is required")
}

// ─── Categories ───────────────────────────────────────────────────────────────

func (h *menuHandler) listCategories(w http.ResponseWriter, r *http.Request) {
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		mw.Error(w, badRequestf("invalid or missing location_id"))
		return
	}
	cats, err := h.menu.ListCategories(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, cats)
}

type createCategoryRequest struct {
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	SortOrder int32  `json:"sort_order"`
}

func (h *menuHandler) createCategory(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Name == "" {
		mw.Error(w, badRequestf("name is required"))
		return
	}
	cat, err := h.menu.CreateCategory(r.Context(), service.CreateCategoryInput{
		LocationID: claims.LocationID,
		Name:       req.Name,
		Icon:       req.Icon,
		SortOrder:  req.SortOrder,
		CreatedBy:  claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, cat)
}

type updateCategoryRequest struct {
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	SortOrder int32  `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

func (h *menuHandler) updateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid category id"))
		return
	}
	var req updateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	cat, err := h.menu.UpdateCategory(r.Context(), service.UpdateCategoryInput{
		ID:        id,
		Name:      req.Name,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, cat)
}

func (h *menuHandler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid category id"))
		return
	}
	if err := h.menu.DeleteCategory(r.Context(), id); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── Products ─────────────────────────────────────────────────────────────────

func (h *menuHandler) listProducts(w http.ResponseWriter, r *http.Request) {
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		mw.Error(w, badRequestf("invalid or missing location_id"))
		return
	}
	activeOnly := r.URL.Query().Get("active") == "true"
	products, err := h.menu.ListProducts(r.Context(), locID, activeOnly)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, products)
}

// listMenu is the POS endpoint: only active, non-stop-listed products.
func (h *menuHandler) listMenu(w http.ResponseWriter, r *http.Request) {
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		mw.Error(w, badRequestf("invalid or missing location_id"))
		return
	}
	products, err := h.menu.ListMenu(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, products)
}

func (h *menuHandler) getProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid product id"))
		return
	}
	product, err := h.menu.GetProduct(r.Context(), id)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, product)
}

type createProductRequest struct {
	CategoryID  string  `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	SKU         *string `json:"sku"`
	BasePrice   float64 `json:"base_price"`
	IsActive    bool    `json:"is_active"`
	ImageURL    *string `json:"image_url"`
	SortOrder   int32   `json:"sort_order"`
}

func (h *menuHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Name == "" {
		mw.Error(w, badRequestf("name is required"))
		return
	}
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		mw.Error(w, badRequestf("invalid category_id"))
		return
	}
	if req.BasePrice < 0 {
		mw.Error(w, badRequestf("base_price must be >= 0"))
		return
	}
	product, err := h.menu.CreateProduct(r.Context(), service.CreateProductInput{
		LocationID:  claims.LocationID,
		CategoryID:  categoryID,
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		BasePrice:   req.BasePrice,
		IsActive:    req.IsActive,
		ImageURL:    req.ImageURL,
		SortOrder:   req.SortOrder,
		CreatedBy:   claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, product)
}

type updateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CategoryID  string  `json:"category_id"`
	BasePrice   float64 `json:"base_price"`
	IsActive    bool    `json:"is_active"`
	ImageURL    *string `json:"image_url"`
	SortOrder   int32   `json:"sort_order"`
	SKU         *string `json:"sku"`
}

func (h *menuHandler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid product id"))
		return
	}
	var req updateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		mw.Error(w, badRequestf("invalid category_id"))
		return
	}
	product, err := h.menu.UpdateProduct(r.Context(), service.UpdateProductInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  categoryID,
		BasePrice:   req.BasePrice,
		IsActive:    req.IsActive,
		ImageURL:    req.ImageURL,
		SortOrder:   req.SortOrder,
		SKU:         req.SKU,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, product)
}

func (h *menuHandler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid product id"))
		return
	}
	if err := h.menu.DeleteProduct(r.Context(), id); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type setStopListRequest struct {
	Stopped bool `json:"stopped"`
}

func (h *menuHandler) setStopList(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid product id"))
		return
	}
	var req setStopListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if err := h.menu.SetStopList(r.Context(), id, req.Stopped); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── Modifiers ────────────────────────────────────────────────────────────────

type createModifierGroupRequest struct {
	Name          string `json:"name"`
	SelectionType string `json:"selection_type"` // "single" | "multi"
	Required      bool   `json:"required"`
	MinSelect     int32  `json:"min_select"`
	MaxSelect     int32  `json:"max_select"`
	SortOrder     int32  `json:"sort_order"`
}

func (h *menuHandler) createModifierGroup(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid product id"))
		return
	}
	var req createModifierGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Name == "" {
		mw.Error(w, badRequestf("name is required"))
		return
	}
	if req.SelectionType == "" {
		req.SelectionType = "single"
	}
	group, err := h.menu.CreateModifierGroup(r.Context(), service.CreateModifierGroupInput{
		ProductID:     productID,
		Name:          req.Name,
		SelectionType: req.SelectionType,
		Required:      req.Required,
		MinSelect:     req.MinSelect,
		MaxSelect:     req.MaxSelect,
		SortOrder:     req.SortOrder,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, group)
}

func (h *menuHandler) deleteModifierGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(chi.URLParam(r, "groupId"))
	if err != nil {
		mw.Error(w, badRequestf("invalid group id"))
		return
	}
	if err := h.menu.DeleteModifierGroup(r.Context(), groupID); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createModifierOptionRequest struct {
	Name               string     `json:"name"`
	PriceDelta         float64    `json:"price_delta"`
	LinkedIngredientID *uuid.UUID `json:"linked_ingredient_id"`
	IngredientQtyDelta float64    `json:"ingredient_qty_delta"`
	SortOrder          int32      `json:"sort_order"`
}

func (h *menuHandler) createModifierOption(w http.ResponseWriter, r *http.Request) {
	groupID, err := uuid.Parse(chi.URLParam(r, "groupId"))
	if err != nil {
		mw.Error(w, badRequestf("invalid group id"))
		return
	}
	var req createModifierOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Name == "" {
		mw.Error(w, badRequestf("name is required"))
		return
	}
	opt, err := h.menu.CreateModifierOption(r.Context(), service.CreateModifierOptionInput{
		GroupID:            groupID,
		Name:               req.Name,
		PriceDelta:         req.PriceDelta,
		LinkedIngredientID: req.LinkedIngredientID,
		IngredientQtyDelta: req.IngredientQtyDelta,
		SortOrder:          req.SortOrder,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, opt)
}

type updateModifierOptionRequest struct {
	Name       string  `json:"name"`
	PriceDelta float64 `json:"price_delta"`
	IsActive   bool    `json:"is_active"`
	SortOrder  int32   `json:"sort_order"`
}

func (h *menuHandler) updateModifierOption(w http.ResponseWriter, r *http.Request) {
	optID, err := uuid.Parse(chi.URLParam(r, "optId"))
	if err != nil {
		mw.Error(w, badRequestf("invalid option id"))
		return
	}
	var req updateModifierOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	opt, err := h.menu.UpdateModifierOption(r.Context(), service.UpdateModifierOptionInput{
		ID:         optID,
		Name:       req.Name,
		PriceDelta: req.PriceDelta,
		IsActive:   req.IsActive,
		SortOrder:  req.SortOrder,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, opt)
}
