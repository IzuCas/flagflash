package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/IzuCas/flagflash/internal/interfaces/http/handler"
	"github.com/IzuCas/flagflash/internal/interfaces/http/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// FlagFlashRouterConfig contains configuration for the FlagFlash router
type FlagFlashRouterConfig struct {
	// Services
	TenantService        *service.TenantService
	ApplicationService   *service.ApplicationService
	EnvironmentService   *service.EnvironmentService
	FeatureFlagService   *service.FeatureFlagService
	TargetingRuleService *service.FeatureFlagService
	APIKeyService        *service.APIKeyService
	EvaluationService    *service.EvaluationService
	AuthService          *service.AuthService
	AuditLogService      *service.AuditLogService
	UsageMetricsService  *service.UsageMetricsService
	UserService          *service.UserService
	// Repositories
	UserRepo repository.UserRepository
	// WebSocket
	WSHandler http.Handler
}

// FlagFlashRouter handles FlagFlash API routing
type FlagFlashRouter struct {
	config *FlagFlashRouterConfig
}

// NewFlagFlashRouter creates a new FlagFlash router
func NewFlagFlashRouter(cfg *FlagFlashRouterConfig) *FlagFlashRouter {
	return &FlagFlashRouter{config: cfg}
}

// SetupRoutes sets up all FlagFlash routes on the given chi router
func (fr *FlagFlashRouter) SetupRoutes(r chi.Router) {
	// FlagFlash API group
	r.Route("/api/v1/flagflash", func(r chi.Router) {
		// Global middleware
		r.Use(chimiddleware.RequestID)
		r.Use(chimiddleware.RealIP)
		r.Use(chimiddleware.Logger)
		r.Use(chimiddleware.Recoverer)
		r.Use(chimiddleware.Compress(5))
		r.Use(corsMiddleware())

		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "healthy", "service": "flagflash"}`))
		})

		// Create Huma API for public routes
		publicAPI := humachi.New(r, huma.DefaultConfig("FlagFlash API", "1.0.0"))
		configureOpenAPI(publicAPI)

		// Register public auth routes (login, register, refresh)
		var authHandler *handler.FlagFlashAuthHandler
		if fr.config.AuthService != nil {
			authHandler = handler.NewFlagFlashAuthHandler(fr.config.AuthService)
			authHandler.RegisterRoutes(publicAPI)
		}

		// Register public invite routes (validate, accept)
		if fr.config.UserService != nil {
			inviteHandler := handler.NewInviteHandler(fr.config.UserService)
			inviteHandler.RegisterRoutes(publicAPI)
		}

		// Protected auth routes (switch-tenant, change-password)
		r.Route("/auth", func(r chi.Router) {
			if fr.config.AuthService != nil {
				r.Use(middleware.FlagFlashJWTAuth(fr.config.AuthService))
			}

			protectedAuthAPI := humachi.New(r, huma.DefaultConfig("FlagFlash Protected Auth API", "1.0.0"))

			if authHandler != nil {
				authHandler.RegisterProtectedRoutes(protectedAuthAPI)
			}
		})

		// SDK routes (API Key auth)
		r.Route("/sdk", func(r chi.Router) {
			if fr.config.APIKeyService != nil {
				r.Use(middleware.APIKeyAuth(fr.config.APIKeyService))
			}

			sdkAPI := humachi.New(r, huma.DefaultConfig("FlagFlash SDK API", "1.0.0"))

			if fr.config.EvaluationService != nil {
				evaluationHandler := handler.NewEvaluationHandler(fr.config.EvaluationService)
				evaluationHandler.RegisterRoutes(sdkAPI)
			}

			// WebSocket endpoint — protected by the same APIKeyAuth middleware above
			if fr.config.WSHandler != nil {
				r.Handle("/ws", fr.config.WSHandler)
			}
		})

		// Dashboard/Management routes
		r.Route("/manage", func(r chi.Router) {
			// Add JWT auth middleware for dashboard
			if fr.config.AuthService != nil {
				r.Use(middleware.FlagFlashJWTAuth(fr.config.AuthService))
			}

			// Register handlers
			dashboardAPI := humachi.New(r, huma.DefaultConfig("FlagFlash Dashboard API", "1.0.0"))

			if fr.config.TenantService != nil {
				tenantHandler := handler.NewTenantHandler(fr.config.TenantService)
				tenantHandler.RegisterRoutes(dashboardAPI)
			}

			if fr.config.ApplicationService != nil {
				appHandler := handler.NewApplicationHandler(fr.config.ApplicationService)
				appHandler.RegisterRoutes(dashboardAPI)
			}

			if fr.config.EnvironmentService != nil {
				envHandler := handler.NewEnvironmentHandler(fr.config.EnvironmentService)
				envHandler.RegisterRoutes(dashboardAPI)
			}

			if fr.config.FeatureFlagService != nil {
				flagHandler := handler.NewFeatureFlagHandler(fr.config.FeatureFlagService)
				flagHandler.RegisterRoutes(dashboardAPI)
			}

			if fr.config.TargetingRuleService != nil {
				ruleHandler := handler.NewTargetingRuleHandler(fr.config.TargetingRuleService)
				ruleHandler.RegisterRoutes(dashboardAPI)
			}

			if fr.config.APIKeyService != nil {
				apiKeyHandler := handler.NewAPIKeyHandler(fr.config.APIKeyService)
				apiKeyHandler.RegisterRoutes(dashboardAPI)
			}

			if fr.config.AuditLogService != nil && fr.config.UserRepo != nil {
				auditLogHandler := handler.NewAuditLogHandler(fr.config.AuditLogService, fr.config.UserRepo)
				auditLogHandler.RegisterRoutes(dashboardAPI)
			}

			if fr.config.UsageMetricsService != nil {
				usageMetricsHandler := handler.NewUsageMetricsHandler(fr.config.UsageMetricsService)
				usageMetricsHandler.RegisterRoutes(dashboardAPI)
			}

			if fr.config.UserService != nil {
				userHandler := handler.NewUserHandler(fr.config.UserService)
				userHandler.RegisterRoutes(dashboardAPI)
			}
		})

	})
}

func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "3600")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func configureOpenAPI(api huma.API) {
	api.OpenAPI().Info.Description = "Feature Flags and Dynamic Configuration Platform API"
	api.OpenAPI().Info.Contact = &huma.Contact{
		Name:  "FlagFlash Team",
		Email: "support@flagflash.io",
	}
}

// FlagFlashServer represents the FlagFlash HTTP server
type FlagFlashServer struct {
	server *http.Server
	router chi.Router
}

// NewFlagFlashServer creates a new FlagFlash HTTP server
func NewFlagFlashServer(port int, router chi.Router) *FlagFlashServer {
	return &FlagFlashServer{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
		router: router,
	}
}

// Start starts the HTTP server
func (s *FlagFlashServer) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *FlagFlashServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Router returns the underlying router
func (s *FlagFlashServer) Router() chi.Router {
	return s.router
}
