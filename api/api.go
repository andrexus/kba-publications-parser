package api

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"

	"context"

	"github.com/andrexus/kba-publications-parser/conf"
	"github.com/andrexus/kba-publications-parser/service"
	"github.com/labstack/echo/middleware"
)

// API is the data holder for the API
type API struct {
	config *conf.Configuration
	log    *logrus.Entry
	echo   *echo.Echo

	// Services used by the API
	kba service.KBAService
}

type ListResponse struct {
	Total int         `json:"total"`
	Items interface{} `json:"items"`
}

type Response struct {
	Message string `json:"message"`
}

// Start will start the API on the specified port
func (api *API) Start() error {
	return api.echo.Start(fmt.Sprintf(":%d", api.config.API.Port))
}

// Stop will shutdown the engine internally
func (api *API) Stop() error {
	logrus.Info("Stopping API server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return api.echo.Shutdown(ctx)
}

// NewAPI will create an api instance that is ready to start
func NewAPI(config *conf.Configuration) *API {
	api := &API{
		config: config,
		log:    logrus.WithField("component", "api"),
	}

	api.kba = &service.KBAServiceImpl{}

	// add the endpoints
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	g := e.Group("/api/v1")

	// UserProfile profile
	g.POST("/upload", api.UploadPDF, middleware.BodyLimit("10M"))

	e.HTTPErrorHandler = api.handleError
	api.echo = e

	return api
}
