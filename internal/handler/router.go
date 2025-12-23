package handler

import (
	"QuanPhotos/internal/config"
	"QuanPhotos/internal/middleware"
	"QuanPhotos/internal/pkg/jwt"
	"QuanPhotos/internal/repository/postgresql/token"
	"QuanPhotos/internal/repository/postgresql/user"
	"QuanPhotos/internal/service/auth"
	userService "QuanPhotos/internal/service/user"
	"QuanPhotos/internal/service/system"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// Router holds the gin router instance
type Router struct {
	engine *gin.Engine
	config *config.Config

	// JWT manager
	jwtManager *jwt.Manager

	// Handlers
	systemHandler *SystemHandler
	authHandler   *AuthHandler
	userHandler   *UserHandler
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.Config, db *sqlx.DB) *Router {
	// Set gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Apply global middleware
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.RequestID())

	// Apply CORS if enabled
	if cfg.CORS.Enabled {
		engine.Use(middleware.CORS(middleware.CORSConfig{
			AllowedOrigins: cfg.CORS.AllowedOrigins,
			AllowedMethods: cfg.CORS.AllowedMethods,
			AllowedHeaders: cfg.CORS.AllowedHeaders,
			MaxAge:         cfg.CORS.MaxAge,
		}))
	}

	// Initialize JWT manager
	jwtManager := jwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
		cfg.JWT.Issuer,
	)

	// Initialize repositories
	userRepo := user.NewUserRepository(db)
	tokenRepo := token.NewTokenRepository(db)

	// Initialize services
	systemService := system.NewService(cfg)
	authService := auth.New(db, userRepo, tokenRepo, jwtManager)
	userSvc := userService.New(userRepo)

	// Initialize handlers
	systemHandler := NewSystemHandler(systemService)
	authHandler := NewAuthHandler(authService)
	userHandler := NewUserHandler(userSvc)

	return &Router{
		engine:        engine,
		config:        cfg,
		jwtManager:    jwtManager,
		systemHandler: systemHandler,
		authHandler:   authHandler,
		userHandler:   userHandler,
	}
}

// Setup sets up all routes
func (r *Router) Setup() {
	// Health check endpoint
	r.engine.GET("/health", r.systemHandler.Health)

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// System routes
		v1.GET("/system/info", r.systemHandler.Info)

		// Auth routes (public)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", r.authHandler.Register)
			authRoutes.POST("/login", r.authHandler.Login)
			authRoutes.POST("/refresh", r.authHandler.Refresh)
			authRoutes.POST("/logout", r.authHandler.Logout)
		}

		// User routes
		users := v1.Group("/users")
		{
			// Public routes
			users.GET("/:id", r.userHandler.GetUser)

			// Protected routes (require authentication)
			users.GET("/me", middleware.Auth(r.jwtManager), r.userHandler.GetCurrentUser)
			users.PUT("/me", middleware.Auth(r.jwtManager), r.userHandler.UpdateCurrentUser)
			users.PUT("/me/password", middleware.Auth(r.jwtManager), r.userHandler.ChangePassword)
		}

		// Photos routes (to be implemented)
		// photos := v1.Group("/photos")

		// Categories routes (to be implemented)
		// categories := v1.Group("/categories")

		// Tags routes (to be implemented)
		// tags := v1.Group("/tags")

		// Tickets routes (to be implemented)
		// tickets := v1.Group("/tickets")

		// Admin routes (to be implemented)
		// admin := v1.Group("/admin")
	}
}

// GetEngine returns the gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Run starts the HTTP server
func (r *Router) Run() error {
	return r.engine.Run(":" + r.config.App.Port)
}
