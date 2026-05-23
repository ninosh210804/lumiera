package handler

import "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/domain"

func badRequestf(msg string) *domain.AppError {
	return domain.NewBadRequest(msg)
}
