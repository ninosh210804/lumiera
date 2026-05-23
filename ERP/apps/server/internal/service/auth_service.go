package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/ninosh210804/lumiera/apps/server/internal/config"
	pgdb "github.com/ninosh210804/lumiera/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
	"github.com/ninosh210804/lumiera/apps/server/internal/token"
)

// AuthService handles authentication and user management.
type AuthService struct {
	q   *pgdb.Queries
	cfg config.JWTConfig
}

// TokenResponse is returned on successful login.
type TokenResponse struct {
	Token     string          `json:"token"`
	ExpiresAt time.Time       `json:"expires_at"`
	User      UserInfo        `json:"user"`
}

// UserInfo is the public user representation embedded in the token response.
type UserInfo struct {
	ID         uuid.UUID       `json:"id"`
	FullName   string          `json:"full_name"`
	Role       domain.RoleType `json:"role"`
	LocationID uuid.UUID       `json:"location_id"`
}

// BaristaItem is shown on the barista PIN-selection screen.
type BaristaItem struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
}

func NewAuthService(q *pgdb.Queries, cfg config.JWTConfig) *AuthService {
	return &AuthService{q: q, cfg: cfg}
}

// LoginEmail authenticates an admin/manager by email + PIN.
func (s *AuthService) LoginEmail(ctx context.Context, email, pin string) (*TokenResponse, error) {
	user, err := s.q.GetUserByEmail(ctx, &email)
	if err != nil {
		// Return generic error to avoid user enumeration
		return nil, domain.ErrInvalidPIN
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PinHash), []byte(pin)); err != nil {
		return nil, domain.ErrInvalidPIN
	}
	userID := uuid.UUID(user.ID.Bytes)
	locID := uuid.UUID(user.DefaultLocationID.Bytes)
	return s.issueToken(userID, locID, domain.RoleType(user.RoleCode), user.FullName)
}

// LoginPIN authenticates a barista by user ID + PIN (no email required).
func (s *AuthService) LoginPIN(ctx context.Context, userID, locationID uuid.UUID, pin string) (*TokenResponse, error) {
	user, err := s.q.GetUserByID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return nil, domain.ErrInvalidPIN
	}
	// Verify the barista belongs to the requested location
	if user.DefaultLocationID != (pgtype.UUID{Bytes: locationID, Valid: true}) {
		return nil, domain.ErrInvalidPIN
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PinHash), []byte(pin)); err != nil {
		return nil, domain.ErrInvalidPIN
	}
	return s.issueToken(uuid.UUID(user.ID.Bytes), locationID, domain.RoleType(user.RoleCode), user.FullName)
}

// ListBaristas returns all active baristas at a location for the PIN screen.
func (s *AuthService) ListBaristas(ctx context.Context, locationID uuid.UUID) ([]BaristaItem, error) {
	rows, err := s.q.ListBaristasByLocation(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list baristas: %w", err)
	}
	items := make([]BaristaItem, len(rows))
	for i, r := range rows {
		items[i] = BaristaItem{
			ID:       uuid.UUID(r.ID.Bytes),
			FullName: r.FullName,
		}
	}
	return items, nil
}

// GetMe returns the public profile for the authenticated user.
func (s *AuthService) GetMe(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	user, err := s.q.GetUserByID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return nil, domain.NewNotFound("user")
	}
	return &UserInfo{
		ID:         uuid.UUID(user.ID.Bytes),
		FullName:   user.FullName,
		Role:       domain.RoleType(user.RoleCode),
		LocationID: uuid.UUID(user.DefaultLocationID.Bytes),
	}, nil
}

func (s *AuthService) issueToken(userID, locationID uuid.UUID, role domain.RoleType, fullName string) (*TokenResponse, error) {
	ttl := s.cfg.TTLHours
	if ttl == 0 {
		ttl = 8 * time.Hour
	}

	tok, exp, err := token.Issue(s.cfg.Secret, ttl, userID, locationID, role)
	if err != nil {
		return nil, domain.NewInternal("token generation failed")
	}

	return &TokenResponse{
		Token:     tok,
		ExpiresAt: exp,
		User: UserInfo{
			ID:         userID,
			FullName:   fullName,
			LocationID: locationID,
			Role:       role,
		},
	}, nil
}

// --- User management ---

// UserService handles user CRUD (admin/manager operations).
type UserService struct {
	q *pgdb.Queries
}

type CreateUserRequest struct {
	LocationID uuid.UUID       `json:"location_id"`
	RoleID     uuid.UUID       `json:"role_id"`
	FullName   string          `json:"full_name"`
	Email      *string         `json:"email"`
	PIN        string          `json:"pin"`
	CreatedBy  uuid.UUID       `json:"-"`
}

type UserListItem struct {
	ID         uuid.UUID       `json:"id"`
	FullName   string          `json:"full_name"`
	Email      *string         `json:"email"`
	Role       string          `json:"role"`
	IsActive   bool            `json:"is_active"`
	LocationID uuid.UUID       `json:"location_id"`
}

func NewUserService(q *pgdb.Queries) *UserService {
	return &UserService{q: q}
}

type RoleDTO struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func (s *UserService) ListRoles(ctx context.Context) ([]RoleDTO, error) {
	rows, err := s.q.ListRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	out := make([]RoleDTO, len(rows))
	for i, r := range rows {
		out[i] = RoleDTO{
			ID:   uuid.UUID(r.ID.Bytes).String(),
			Code: r.Code,
			Name: r.Name,
		}
	}
	return out, nil
}

// ListByLocation returns all users for a given location.
func (s *UserService) ListByLocation(ctx context.Context, locationID uuid.UUID) ([]UserListItem, error) {
	rows, err := s.q.ListUsersByLocation(ctx, pgtype.UUID{Bytes: locationID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	items := make([]UserListItem, len(rows))
	for i, r := range rows {
		items[i] = UserListItem{
			ID:         uuid.UUID(r.ID.Bytes),
			FullName:   r.FullName,
			Email:      r.Email,
			Role:       r.RoleCode,
			IsActive:   r.IsActive,
			LocationID: uuid.UUID(r.DefaultLocationID.Bytes),
		}
	}
	return items, nil
}

// Create adds a new user. PIN is bcrypt-hashed before storage.
func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (*UserListItem, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.NewInternal("hashing PIN failed")
	}
	user, err := s.q.CreateUser(ctx, pgdb.CreateUserParams{
		DefaultLocationID: pgtype.UUID{Bytes: req.LocationID, Valid: true},
		RoleID:            pgtype.UUID{Bytes: req.RoleID, Valid: true},
		FullName:          req.FullName,
		Email:             req.Email,
		PinHash:           string(hash),
		CreatedBy:         pgtype.UUID{Bytes: req.CreatedBy, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &UserListItem{
		ID:         uuid.UUID(user.ID.Bytes),
		FullName:   user.FullName,
		Email:      user.Email,
		IsActive:   user.IsActive,
		LocationID: uuid.UUID(user.DefaultLocationID.Bytes),
	}, nil
}

// UpdatePIN replaces a user's PIN hash.
func (s *UserService) UpdatePIN(ctx context.Context, userID uuid.UUID, newPIN string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPIN), bcrypt.DefaultCost)
	if err != nil {
		return domain.NewInternal("hashing PIN failed")
	}
	return s.q.UpdateUserPIN(ctx, pgdb.UpdateUserPINParams{
		ID:      pgtype.UUID{Bytes: userID, Valid: true},
		PinHash: string(hash),
	})
}

// Deactivate soft-deletes a user (sets is_active = false).
func (s *UserService) Deactivate(ctx context.Context, userID uuid.UUID) error {
	return s.q.DeactivateUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
}
