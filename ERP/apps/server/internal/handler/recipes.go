package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	mw "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/middleware"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/service"
)

type recipesHandler struct {
	recipes *service.RecipeService
}

// GET /api/v1/recipes?location_id=
func (h *recipesHandler) list(w http.ResponseWriter, r *http.Request) {
	locID, err := locationFromCtxOrQuery(r)
	if err != nil {
		mw.Error(w, badRequestf("invalid or missing location_id"))
		return
	}
	items, err := h.recipes.List(r.Context(), locID)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, items)
}

// GET /api/v1/recipes/{id}
func (h *recipesHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid recipe id"))
		return
	}
	recipe, err := h.recipes.Get(r.Context(), id)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, recipe)
}

// POST /api/v1/recipes
type createRecipeRequest struct {
	Name       string  `json:"name"`
	RecipeType string  `json:"recipe_type"`
	YieldQty   float64 `json:"yield_qty"`
	YieldUnit  string  `json:"yield_unit"`
	Notes      string  `json:"notes"`
}

func (h *recipesHandler) create(w http.ResponseWriter, r *http.Request) {
	claims, _ := mw.ClaimsFrom(r.Context())
	locID, _ := locationFromCtxOrQuery(r)
	if locID == (uuid.UUID{}) {
		locID = claims.LocationID
	}
	var req createRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Name == "" {
		mw.Error(w, badRequestf("name is required"))
		return
	}
	recipe, err := h.recipes.Create(r.Context(), service.CreateRecipeInput{
		LocationID: locID,
		Name:       req.Name,
		RecipeType: req.RecipeType,
		YieldQty:   req.YieldQty,
		YieldUnit:  req.YieldUnit,
		Notes:      req.Notes,
		CreatedBy:  claims.UserID,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, recipe)
}

// PUT /api/v1/recipes/{id}
type updateRecipeRequest struct {
	Name       string  `json:"name"`
	RecipeType string  `json:"recipe_type"`
	YieldQty   float64 `json:"yield_qty"`
	YieldUnit  string  `json:"yield_unit"`
	Notes      string  `json:"notes"`
}

func (h *recipesHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid recipe id"))
		return
	}
	var req updateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	recipe, err := h.recipes.Update(r.Context(), service.UpdateRecipeInput{
		ID:         id,
		Name:       req.Name,
		RecipeType: req.RecipeType,
		YieldQty:   req.YieldQty,
		YieldUnit:  req.YieldUnit,
		Notes:      req.Notes,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, recipe)
}

// DELETE /api/v1/recipes/{id}
func (h *recipesHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid recipe id"))
		return
	}
	if err := h.recipes.Delete(r.Context(), id); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /api/v1/recipes/{id}/link — link recipe to a product
type linkProductRequest struct {
	ProductID string `json:"product_id"`
}

func (h *recipesHandler) linkProduct(w http.ResponseWriter, r *http.Request) {
	recipeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid recipe id"))
		return
	}
	var req linkProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		mw.Error(w, badRequestf("invalid product_id"))
		return
	}
	if err := h.recipes.LinkToProduct(r.Context(), productID, recipeID); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DELETE /api/v1/recipes/{id}/link/{productId}
func (h *recipesHandler) unlinkProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		mw.Error(w, badRequestf("invalid product_id"))
		return
	}
	if err := h.recipes.UnlinkFromProduct(r.Context(), productID); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ─── Recipe items ─────────────────────────────────────────────────────────────

// POST /api/v1/recipes/{id}/items
type addRecipeItemRequest struct {
	IngredientID *string `json:"ingredient_id"`
	SubRecipeID  *string `json:"sub_recipe_id"`
	Qty          float64 `json:"qty"`
	Unit         string  `json:"unit"`
	SortOrder    int32   `json:"sort_order"`
}

func (h *recipesHandler) addItem(w http.ResponseWriter, r *http.Request) {
	recipeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		mw.Error(w, badRequestf("invalid recipe id"))
		return
	}
	var req addRecipeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	if req.Qty <= 0 {
		mw.Error(w, badRequestf("qty must be > 0"))
		return
	}
	if req.Unit == "" {
		mw.Error(w, badRequestf("unit is required"))
		return
	}

	in := service.AddRecipeItemInput{
		RecipeID:  recipeID,
		Qty:       req.Qty,
		Unit:      req.Unit,
		SortOrder: req.SortOrder,
	}
	if req.IngredientID != nil {
		id, err := uuid.Parse(*req.IngredientID)
		if err != nil {
			mw.Error(w, badRequestf("invalid ingredient_id"))
			return
		}
		in.IngredientID = &id
	}
	if req.SubRecipeID != nil {
		id, err := uuid.Parse(*req.SubRecipeID)
		if err != nil {
			mw.Error(w, badRequestf("invalid sub_recipe_id"))
			return
		}
		in.SubRecipeID = &id
	}

	item, err := h.recipes.AddItem(r.Context(), in)
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusCreated, item)
}

// PUT /api/v1/recipes/{id}/items/{itemId}
type updateRecipeItemRequest struct {
	Qty       float64 `json:"qty"`
	Unit      string  `json:"unit"`
	SortOrder int32   `json:"sort_order"`
}

func (h *recipesHandler) updateItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		mw.Error(w, badRequestf("invalid item id"))
		return
	}
	var req updateRecipeItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mw.Error(w, badRequestf("invalid JSON"))
		return
	}
	item, err := h.recipes.UpdateItem(r.Context(), service.UpdateRecipeItemInput{
		ID:        itemID,
		Qty:       req.Qty,
		Unit:      req.Unit,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, item)
}

// DELETE /api/v1/recipes/{id}/items/{itemId}
func (h *recipesHandler) removeItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		mw.Error(w, badRequestf("invalid item id"))
		return
	}
	if err := h.recipes.RemoveItem(r.Context(), itemID); err != nil {
		mw.Error(w, err)
		return
	}
	mw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
