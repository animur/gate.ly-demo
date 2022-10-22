package controller

import (
	"context"
	"fmt"
	"net/http"

	"gately/internal/config"
	"gately/internal/multicache"
	"gately/internal/service"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type AppController struct {
	uss *service.UrlShorteningService
}

type (
	UrlMappingResponse struct {
		LongUrl  int    `json:"long_url"`
		ShortUrl string `json:"short_curl"`
		Ts       string `json:"Time"`
	}

	UrlMappingRequest struct {
		LongUrl int `json:"long_url"`
	}
)

func New(cfg config.AppConfig) *AppController {

	// Instantiate our multicache
	// This follows a dual layered caching strategy
	cache := multicache.New(cfg)

	// Instantiate MongoDB Client
	// MongoDB is our source of truth for all URL mappings
	// This is a read heavy application and MongoDB is best suited for read heavy apps
	mongoURI := fmt.Sprintf("%s://%s", "mongodb", cfg.MongoHost)
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		// Ok to panic as we are still in application bootstrap
		panic(err)
	}

	// Try to ping MongoDB to test connectivity
	if err := mongoClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		// Ok to panic as we are still in application bootstrap
		panic(err)
	}

	urlServ := service.New(
		service.WithMultiCache(cache),
		service.WithMongoDB(mongoClient),
	)
	fmt.Print("Successfully connected to MongoDB and Redis")
	return &AppController{uss: urlServ}
}

func (ctrlr *AppController) CreateUrlMapping(c echo.Context) error {

	resp := &UrlMappingResponse{}

	// service call
	if err := c.Bind(resp); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, resp)
}

func (ctrlr *AppController) DeleteUrlMapping(c echo.Context) error {

	// urlId := (c.Param("urlId"))

	return c.NoContent(http.StatusNoContent)
}

func (ctrlr *AppController) RedirectUrl(c echo.Context) error {

	// urlId := c.Param("urlId")
	longUrl := ""
	// Redirect to the original URL
	return c.Redirect(http.StatusSeeOther, longUrl)
}
