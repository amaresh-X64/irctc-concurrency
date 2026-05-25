package main

import (
	"log"

	"gin-booking/config"
	"gin-booking/internal/booking"
	"gin-booking/internal/middleware"
	appdb "gin-booking/pkg/db"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db := appdb.Connect(cfg.DatabaseURL)
	defer db.Close()

	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "IRCTC Booking Service"})
	})

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware()) // 🔒 all booking routes protected

	bookingService := booking.NewService(db)
	bookingController := booking.NewController(bookingService)
	bookingController.RegisterRoutes(api)

	log.Printf("Gin Booking Service running on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}