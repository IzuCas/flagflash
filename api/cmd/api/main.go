package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/infrastructure/config"
	"github.com/IzuCas/flagflash/internal/infrastructure/email"
	"github.com/IzuCas/flagflash/internal/infrastructure/postgres"
	"github.com/IzuCas/flagflash/internal/infrastructure/redis"
	wsinfra "github.com/IzuCas/flagflash/internal/infrastructure/websocket"
	httpRouter "github.com/IzuCas/flagflash/internal/interfaces/http"
	"github.com/IzuCas/flagflash/pkg/logger"
	"github.com/IzuCas/flagflash/pkg/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	if err := logger.Init(env); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// Startup security validation
	if cfg.JWT.Secret == "" {
		panic("FATAL: JWT_SECRET environment variable must be set before starting the server")
	}
	if len(cfg.JWT.Secret) < 32 {
		logger.Warn("SECURITY WARNING: JWT_SECRET is shorter than 32 characters — use a cryptographically random secret in production")
	}

	logger.Info("Starting FlagFlash API",
		logger.String("env", env),
		logger.String("version", "1.0.0"),
	)

	// Initialize PostgreSQL connection
	dbConfig := &postgres.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}
	db, err := postgres.NewConnection(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", logger.Err(err))
	}
	defer db.Close()
	logger.Info("PostgreSQL connected",
		logger.String("host", cfg.Database.Host),
		logger.String("database", cfg.Database.DBName),
	)

	// Initialize Redis client
	redisConfig := &redis.Config{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	redisClient, err := redis.NewClient(redisConfig)
	if err != nil {
		logger.Warn("Failed to connect to Redis - caching disabled", logger.Err(err))
		redisClient = nil
	} else {
		logger.Info("Redis connected",
			logger.String("host", cfg.Redis.Host),
			logger.Int("db", cfg.Redis.DB),
		)
	}

	// Initialize repositories
	tenantRepo := postgres.NewTenantRepo(db)
	appRepo := postgres.NewApplicationRepo(db)
	envRepo := postgres.NewEnvironmentRepository(db.DB)
	flagRepo := postgres.NewFeatureFlagRepository(db.DB)
	targetingRepo := postgres.NewTargetingRuleRepository(db.DB)
	auditRepo := postgres.NewAuditLogRepo(db)
	userRepo := postgres.NewUserRepository(db.DB)
	userMembershipRepo := postgres.NewUserTenantMembershipRepository(db.DB)
	apiKeyRepo := postgres.NewAPIKeyRepo(db)
	evalEventRepo := postgres.NewEvaluationEventRepository(db.DB)
	inviteTokenRepo := postgres.NewInviteTokenRepository(db.DB)

	// Initialize Redis cache (if available)
	var flagCache service.FlagCache
	var flagPublisher service.FlagPublisher
	var apiKeyCache *redis.APIKeyCache
	var appCache *redis.ApplicationCache
	var envCache *redis.EnvironmentCache
	if redisClient != nil {
		flagCache = redis.NewFlagCache(redisClient)
		flagPublisher = redis.NewFlagPublisher(redisClient)
		apiKeyCache = redis.NewAPIKeyCache(redisClient)
		appCache = redis.NewApplicationCache(redisClient)
		envCache = redis.NewEnvironmentCache(redisClient)
		logger.Info("Redis caches initialized (API Keys, Applications, Environments, Flags)")
	}

	// Initialize WebSocket hub
	var pubsub *redis.PubSub
	if redisClient != nil {
		pubsub = redis.NewPubSub(redisClient)
	}
	wsHub := wsinfra.NewHub(pubsub)
	go wsHub.Run(context.Background())
	logger.Info("WebSocket hub started")

	// Initialize services
	tenantService := service.NewTenantService(tenantRepo, userMembershipRepo, auditRepo)
	appService := service.NewApplicationService(appRepo, envRepo, auditRepo, appCache)
	envService := service.NewEnvironmentService(envRepo, appRepo, auditRepo, flagRepo, envCache)
	flagService := service.NewFeatureFlagService(flagRepo, targetingRepo, envRepo, auditRepo, flagCache, flagPublisher)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, tenantRepo, apiKeyCache)
	evaluationService := service.NewEvaluationService(flagRepo, targetingRepo, flagCache)
	authService := service.NewAuthService(userRepo, tenantRepo, userMembershipRepo, auditRepo, cfg.JWT.Secret, cfg.JWT.Expiration)
	auditLogService := service.NewAuditLogService(auditRepo)
	usageMetricsService := service.NewUsageMetricsService(evalEventRepo)

	// Initialize email service
	emailService := email.NewService(&email.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
	})
	if emailService.IsConfigured() {
		logger.Info("SMTP email service configured", logger.String("host", cfg.SMTP.Host))
	} else {
		logger.Warn("SMTP not configured - invite emails will not be sent. Set SMTP_HOST, SMTP_FROM etc.")
	}

	userService := service.NewUserService(userRepo, userMembershipRepo, tenantRepo, auditRepo, inviteTokenRepo, emailService, cfg.AppURL)

	logger.Info("Services initialized")

	// Create Chi router
	r := chi.NewMux()

	// Configure CORS allowed origins from environment variable.
	// Set CORS_ALLOWED_ORIGINS to a comma-separated list of origins, e.g.:
	//   CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
	// Defaults to "*" in development and warns in production.
	allowedOrigins := []string{"*"}
	if raw := os.Getenv("CORS_ALLOWED_ORIGINS"); raw != "" {
		allowedOrigins = strings.Split(raw, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	} else if env == "production" {
		logger.Warn("SECURITY WARNING: CORS_ALLOWED_ORIGINS not set in production — defaulting to '*'. " +
			"Set CORS_ALLOWED_ORIGINS to restrict cross-origin requests.")
	}

	// Add middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(middleware.ZapLogger())
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	// Global rate limit: 300 requests/min per IP across all endpoints
	r.Use(middleware.RateLimit(300))

	// Initialize FlagFlash router
	flagflashRouter := httpRouter.NewFlagFlashRouter(&httpRouter.FlagFlashRouterConfig{
		TenantService:        tenantService,
		ApplicationService:   appService,
		EnvironmentService:   envService,
		FeatureFlagService:   flagService,
		TargetingRuleService: flagService,
		APIKeyService:        apiKeyService,
		EvaluationService:    evaluationService,
		AuthService:          authService,
		AuditLogService:      auditLogService,
		UsageMetricsService:  usageMetricsService,
		UserService:          userService,
		UserRepo:             userRepo,
		AppURL:               cfg.AppURL,
		WSHandler:            wsHub.NewSDKHandler(),
	})

	// Setup FlagFlash routes
	flagflashRouter.SetupRoutes(r)

	// Health check at root
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "flagflash-api"}`))
	})

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down...")
		cancel()

		// Give ongoing requests time to complete
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		_ = shutdownCtx
	}()
	_ = ctx

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "9001"
	}
	logger.Info("Server configuration",
		logger.String("address", ":"+port),
		logger.String("docs", "http://localhost:"+port+"/api/v1/flagflash/docs"),
	)
	logger.Info("FlagFlash API endpoints",
		logger.String("health", "http://localhost:"+port+"/health"),
		logger.String("api", "http://localhost:"+port+"/api/v1/flagflash"),
		logger.String("sdk", "http://localhost:"+port+"/api/v1/flagflash/sdk"),
		logger.String("manage", "http://localhost:"+port+"/api/v1/flagflash/manage"),
	)

	logger.Info("Starting HTTP server", logger.String("port", port))
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Fatal("Failed to start server", logger.Err(err))
	}
}
