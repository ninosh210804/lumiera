package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ninosh210804/lumiera/apps/server/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileWithoutPhoneReturnsEmptyProfile(t *testing.T) {
	h := clientsHandler{orders: &service.OrderService{}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clients/profile", nil)
	rr := httptest.NewRecorder()

	h.profile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var payload struct {
		Data struct {
			Phone          string  `json:"phone"`
			CustomerFound  bool    `json:"customer_found"`
			Balance        float64 `json:"balance"`
			FreeDrinksLeft int     `json:"free_drinks_left"`
			CoffeePunches  int     `json:"coffee_punches"`
		} `json:"data"`
	}

	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &payload))
	assert.Equal(t, "", payload.Data.Phone)
	assert.False(t, payload.Data.CustomerFound)
	assert.Zero(t, payload.Data.Balance)
	assert.Zero(t, payload.Data.FreeDrinksLeft)
	assert.Zero(t, payload.Data.CoffeePunches)
}
