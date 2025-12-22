package handler

import (
	"QuanPhotos/internal/config"
	"QuanPhotos/internal/middleware"
	"QuanPhotos/internal/service/system"

	"github.com/gin-gonic/gin"
)

// Router holds the gin router instance
type Router struct {
	engine *gin.Engine
	config *config.Config

	// Handlers
	systemHandler *SystemHandler
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.Config) *Router {
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

	// Initialize services
	systemService := system.NewService(cfg)

	// Initialize handlers
	systemHandler := NewSystemHandler(systemService)

	return &Router{
		engine:        engine,
		config:        cfg,
		systemHandler: systemHandler,
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

		// Auth routes (to be implemented)
		// auth := v1.Group("/auth")
		// {
		//     auth.POST("/register", authHandler.Register)
		//     auth.POST("/login", authHandler.Login)
		//     auth.POST("/refresh", authHandler.Refresh)
		//     auth.POST("/logout", authHandler.Logout)
		// }

		// User routes (to be implemented)
		// users := v1.Group("/users")
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
