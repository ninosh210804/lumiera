package handler

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/config"
	pgdb "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/db/postgres/generated"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/domain"
	mw "github.com/ninosh210804/coffeeshop-erp/apps/server/internal/middleware"
	"github.com/ninosh210804/coffeeshop-erp/apps/server/internal/service"
)

type healthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

// NewRouter wires all HTTP handlers and returns a chi router.
// acctSvc is optional (nil when accounting is disabled).
func NewRouter(pool *pgxpool.Pool, cfg *config.Config, acctSvc AccountingServicer) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthHandler(pool))
	r.Get("/version", versionHandler())

	r.Route("/api/v1", func(r chi.Router) {
		// Build shared DB queries (nil-safe: only used when pool != nil)
		var q *pgdb.Queries
		if pool != nil {
			q = pgdb.New(pool)
		}

		authSvc := service.NewAuthService(q, cfg.JWT)
		userSvc := service.NewUserService(q)
		menuSvc := service.NewMenuService(q)
		ingSvc := service.NewIngredientService(q)
		recipeSvc := service.NewRecipeService(q)
		orderSvc := service.NewOrderService(pool, q)
		stockSvc := service.NewStockService(pool, q)
		shiftSvc := service.NewShiftService(q)
		analyticsSvc := service.NewAnalyticsService(q)
		syncSvc := service.NewSyncService(q)

		ah := &authHandler{auth: authSvc}
		uh := &usersHandler{users: userSvc}
		mh := &menuHandler{menu: menuSvc}
		ih := &inventoryHandler{ing: ingSvc}
		rh := &recipesHandler{recipes: recipeSvc}
		oh := &ordersHandler{orders: orderSvc}
		sh := &stockHandler{stock: stockSvc}
		suph := &suppliersHandler{stock: stockSvc}
		shiftH := &shiftsHandler{shifts: shiftSvc}
		analyticsH := &analyticsHandler{analytics: analyticsSvc}
		syncH := &syncHandler{sync: syncSvc}

		// Public auth routes (no JWT required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", ah.loginEmail)
			r.Post("/pin", ah.loginPIN)
			r.Get("/baristas", ah.listBaristas)

			// Protected: requires valid JWT
			r.Group(func(r chi.Router) {
				r.Use(mw.Authenticate(cfg.JWT.Secret))
				r.Use(mw.RequireAuth)
				r.Get("/me", ah.me)
			})
		})

		// Public menu endpoint for POS (active products only, no auth required)
		r.Get("/menu", mh.listMenu)

		// Categories (admin/manager write, all authenticated read)
		r.Route("/categories", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Get("/", mh.listCategories)
			r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/", mh.createCategory)
			r.Route("/{id}", func(r chi.Router) {
				r.Use(mw.RequireRole(domain.RoleAdmin, domain.RoleManager))
				r.Put("/", mh.updateCategory)
				r.Delete("/", mh.deleteCategory)
			})
		})

		// Products
		r.Route("/products", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Get("/", mh.listProducts)
			r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/", mh.createProduct)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", mh.getProduct)
				r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Put("/", mh.updateProduct)
				r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Delete("/", mh.deleteProduct)
				// Stop-list toggle: barista can also call this
				r.Post("/stop-list", mh.setStopList)
				// Modifier groups
				r.Route("/modifiers", func(r chi.Router) {
					r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/", mh.createModifierGroup)
					r.Route("/{groupId}", func(r chi.Router) {
						r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Delete("/", mh.deleteModifierGroup)
						r.Route("/options", func(r chi.Router) {
							r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/", mh.createModifierOption)
							r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Put("/{optId}", mh.updateModifierOption)
						})
					})
				})
			})
		})

		// Ingredients (admin/manager write, all authenticated read)
		r.Route("/ingredients", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Get("/", ih.listIngredients)
			r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/", ih.createIngredient)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", ih.getIngredient)
				r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Put("/", ih.updateIngredient)
				r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Delete("/", ih.deleteIngredient)
			})
		})

		// Recipes / tech cards
		r.Route("/recipes", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Use(mw.RequireRole(domain.RoleAdmin, domain.RoleManager))
			r.Get("/", rh.list)
			r.Post("/", rh.create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", rh.get)
				r.Put("/", rh.update)
				r.Delete("/", rh.delete)
				r.Post("/link", rh.linkProduct)
				r.Delete("/link/{productId}", rh.unlinkProduct)
				r.Post("/items", rh.addItem)
				r.Put("/items/{itemId}", rh.updateItem)
				r.Delete("/items/{itemId}", rh.removeItem)
			})
		})

		// Stock operations (write-off, receive, inventory count)
		r.Route("/stock", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Get("/movements", sh.listMovements)
			r.Get("/counts", sh.listCounts)
			r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/write-off", sh.writeOff)
			r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/receive", sh.receiveStock)
			r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/count", sh.createCount)
		})

		// Suppliers
		r.Route("/suppliers", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Use(mw.RequireRole(domain.RoleAdmin, domain.RoleManager))
			r.Get("/", suph.list)
			r.Post("/", suph.create)
		})

		// Orders / POS
		r.Route("/orders", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Get("/payment-methods", oh.listPaymentMethods)
			r.Get("/", oh.list)
			r.Post("/", oh.create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", oh.get)
				r.Get("/receipt", oh.receipt)
				r.With(mw.RequireRole(domain.RoleAdmin, domain.RoleManager)).Post("/cancel", oh.cancel)
			})
		})

		// Shifts
		r.Route("/shifts", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Get("/", shiftH.list)
			r.Post("/open", shiftH.open)
			r.Get("/active", shiftH.active)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", shiftH.get)
				r.Post("/close", shiftH.close)
				r.Get("/cash", shiftH.listCash)
				r.Post("/cash", shiftH.addCash)
			})
		})

		// Analytics / dashboard (admin + manager)
		r.Route("/analytics", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Use(mw.RequireRole(domain.RoleAdmin, domain.RoleManager))
			r.Get("/heatmap", analyticsH.heatmap)
			r.Get("/products", analyticsH.productABC)
			r.Get("/revenue", analyticsH.dailyRevenue)
			r.Get("/baristas", analyticsH.baristaStats)
			r.Post("/refresh", analyticsH.refreshViews)
		})

		// Sync (all authenticated devices)
		r.Route("/sync", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Post("/register", syncH.register)
			r.Post("/push", syncH.push)
			r.Get("/pull", syncH.pull)
			r.Get("/snapshot", syncH.snapshot)
		})

		// Accounting (optional module, admin only)
		if acctSvc != nil {
			acctH := &accountingHandler{acct: acctSvc}
			r.Route("/accounting", func(r chi.Router) {
				r.Use(mw.Authenticate(cfg.JWT.Secret))
				r.Use(mw.RequireAuth)
				r.Use(mw.RequireRole(domain.RoleAdmin))
				r.Get("/categories", acctH.listCategories)
				r.Get("/expenses", acctH.listExpenses)
				r.Post("/expenses", acctH.createExpense)
				r.Get("/journal", acctH.listJournal)
			})
		}

		// Roles (admin + manager, read-only)
		r.Group(func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Use(mw.RequireRole(domain.RoleAdmin, domain.RoleManager))
			r.Get("/roles", uh.listRoles)
		})

		// User management (admin + manager only)
		r.Route("/users", func(r chi.Router) {
			r.Use(mw.Authenticate(cfg.JWT.Secret))
			r.Use(mw.RequireAuth)
			r.Use(mw.RequireRole(domain.RoleAdmin, domain.RoleManager))

			r.Get("/", uh.list)
			r.Post("/", uh.create)

			r.Route("/{id}", func(r chi.Router) {
				r.Put("/pin", uh.updatePIN)
				r.Delete("/", mw.RequireRole(domain.RoleAdmin)(http.HandlerFunc(uh.deactivate)).ServeHTTP)
			})
		})
	})

	return r
}

func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool != nil {
			if err := pool.Ping(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"status": "unhealthy",
					"error":  "database unreachable",
				})
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(healthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Version:   buildVersion(),
		})
	}
}

func versionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"version": buildVersion(),
			"go":      goVersion(),
		})
	}
}

func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "dev"
}

func goVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.GoVersion
	}
	return "unknown"
}
