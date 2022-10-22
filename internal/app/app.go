package app

import (
	"fmt"

	"gately/internal/config"
	"gately/internal/controller"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Run(cfg config.AppConfig) {

	ctrlr := controller.New(cfg)

	if ctrlr == nil {
		panic("Unable to instantiate Application Controller")
	}
	e := echo.New()

	// Try to recover from all panics
	e.Use(middleware.Recover())

	// Create a short url
	e.POST("/api/v1/urls", ctrlr.CreateUrlMapping)
	// Redirect to a real URL given a shortURL
	e.GET("/:urlID", ctrlr.RedirectUrl)
	// Delete a mapped URL
	e.DELETE("/api/v1/urls/:urlId", ctrlr.DeleteUrlMapping)

	// Start server
	address := fmt.Sprintf(":%s", cfg.Port)
	e.Logger.Fatal(e.Start(address))
}
