package handler

import "github.com/ninosh210804/lumiera/apps/server/internal/domain"

func badRequestf(msg string) *domain.AppError {
	return domain.NewBadRequest(msg)
}
