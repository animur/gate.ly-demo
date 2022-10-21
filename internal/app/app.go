package app

import (
	"time"

	"gately/internal/controller"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Port      string `mapstructure:"port"`
	RedisHost string `mapstructure:"redis-host"`
	RedisPass string `mapstructure:"redis-pass"`
	MongoHost string `mapstructure:"mongo-host"`
	MongoUser string `mapstructure:"mongo-user"`
	MongoPass string `mapstructure:"mongo-pass"`
}

// ZapLogger is an example of echo middleware that logs requests using logger "zap"
func ZapLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			fields := []zapcore.Field{
				zap.Int("status", res.Status),
				zap.String("latency", time.Since(start).String()),
				zap.String("id", id),
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.String("host", req.Host),
				zap.String("remote_ip", c.RealIP()),
			}

			n := res.Status
			switch {
			case n >= 500:
				log.Error("Server error", fields...)
			case n >= 400:
				log.Warn("Client error", fields...)
			case n >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}

			return nil
		}
	}
}

func Run(cfg Config) {

	ctrlr := controller.New(cfg)

	e := echo.New()

	// Use Zap logger instead of Echo framework's default
	e.Use(ZapLogger(&zap.Logger{}))
	// Try to recover from all panics
	e.Use(middleware.Recover())

	// Create a short url
	e.POST("/api/v1/urls", ctrlr.CreateUrlMapping)
	// Redirect to a real URL given a shortURL
	e.GET("/:urlID", ctrlr.RedirectUrl)
	// Delete a mapped URL
	e.DELETE("/api/v1/urls/:urlId", ctrlr.DeleteUrlMapping)

	// Start server
	e.Logger.Fatal(e.Start(":80"))
}
