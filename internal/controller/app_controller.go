package controller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"gately/internal/config"
	"gately/internal/dal"
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
		LongUrl  string `json:"long_url"`
		ShortUrl string `json:"short_url"`
		Ts       string `json:"time"`
	}

	UrlMappingRequest struct {
		LongUrl string `json:"long_url"`
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
	credential := options.Credential{
		Username: cfg.MongoUser,
		Password: cfg.MongoPass,
	}
	mongoClient, err := mongo.Connect(context.TODO(),
		options.Client().ApplyURI(mongoURI).SetAuth(credential),
	)
	if err != nil {
		// Ok to panic as we are still in application bootstrap
		panic(err)
	}

	// Try to ping MongoDB to test connectivity
	if err := mongoClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		// Ok to panic as we are still in application bootstrap
		panic(err)
	}

	urlStore := dal.New(dal.WithMongoDB(mongoClient), dal.WithDatabase(cfg.MongoDbName), dal.WithTable(cfg.MongoCollectionName))

	urlServ := service.New(
		service.WithMultiCache(cache),
		service.WithUrlStore(urlStore),
	)
	fmt.Print("Successfully connected to MongoDB and Redis")
	return &AppController{uss: urlServ}
}

func (ctrlr *AppController) CreateUrlMapping(c echo.Context) error {

	var req UrlMappingRequest
	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	sanitized, valid := ctrlr.uss.CheckAndSanitizeUrl(req.LongUrl)

	if !valid {
		return c.String(http.StatusBadRequest, "Malformed URL")
	}
	mapped, err := ctrlr.uss.CreateUrlMapping(c.Request().Context(), sanitized)

	if err != nil || mapped == "" {

		if errors.Is(err, dal.ErrUrlEntryAlreadyExists) {
			log.Printf("URL already exists Mapped = %s .Err = %v", mapped, err)
			return c.String(http.StatusBadRequest, err.Error())
		}

		log.Printf("Unable to create a URL mapping. Mapped = %s .Err = %v", mapped, err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	resp := &UrlMappingResponse{
		LongUrl:  req.LongUrl,
		ShortUrl: mapped,
		Ts:       time.Now().String(),
	}

	log.Printf("Successfully created URL mapping : %+v ", resp)

	return c.JSONPretty(http.StatusCreated, resp, "  ")

}

func (ctrlr *AppController) DeleteUrlMapping(c echo.Context) error {

	urlId := c.Param("urlId")

	err := ctrlr.uss.DeleteUrlMapping(c.Request().Context(), urlId)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Unable to delete: %v", err))
	}
	return c.NoContent(http.StatusNoContent)
}

func (ctrlr *AppController) RedirectUrl(c echo.Context) error {

	urlId := c.Param("urlId")
	longUrl, err := ctrlr.uss.RedirectUrl(c.Request().Context(), urlId)

	if longUrl == "" || err != nil {

		switch err {

		case dal.ErrUrlEntryNotFound:
			return c.JSON(http.StatusNotFound, fmt.Sprintf("No short URL found for %s", urlId))
		default:
			return c.JSON(http.StatusInternalServerError, "Internal Server error")
		}

	}
	// Redirect to the original URL
	return c.Redirect(http.StatusSeeOther, longUrl)
}

func (ctrlr *AppController) FetchUsageMetrics(c echo.Context) error {

	return nil
}
