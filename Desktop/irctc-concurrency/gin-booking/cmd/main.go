package main

import (
	"log"

	"gin-booking/config"
	"gin-booking/internal/booking"
	"gin-booking/internal/middleware"
	"gin-booking/pkg/telemetry"
	appdb "gin-booking/pkg/db"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	// ── Telemetry: must be first, before any spans are created ──
	shutdownTracer := telemetry.Init()
	defer shutdownTracer()

	cfg := config.Load()
	db := appdb.Connect(cfg.DatabaseURL)
	defer db.Close()

	r := gin.Default()
	r.Use(middleware.CORS())

	// ── OTel middleware: wraps every HTTP handler in a span ──
	// This automatically gives you span per route with status code,
	// HTTP method, URL, and duration — zero extra code needed per handler.
	r.Use(otelgin.Middleware("gin-booking"))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "IRCTC Booking Service"})
	})

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())

	bookingService := booking.NewService(db)
	bookingController := booking.NewController(bookingService)
	bookingController.RegisterRoutes(api)

	log.Printf("Gin Booking Service running on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
