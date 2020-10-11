package main

import (
	"context"
	"github.com/Octops/agones-relay-http/internal/runtime"
	"github.com/Octops/agones-relay-http/pkg/broker"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"strings"
)

/*
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"username":"xyz","password":"xyz"}' \
  http://localhost:8090
*/

var addr = ":8090"

func main() {
	e := echo.New()
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/webhook", func(c echo.Context) error {
		var payload broker.Payload
		if err := c.Bind(&payload); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "request body does not contain a valid payload")
		}

		log.Printf("webhook received: %s/%s", strings.ToLower(payload.Body.Header.Headers["event_source"]), strings.ToLower(payload.Body.Header.Headers["event_type"]))
		return c.String(http.StatusCreated, "OK")
	})

	e.PUT("/webhook", func(c echo.Context) error {
		var payload broker.Payload
		if err := c.Bind(&payload); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "request body does not contain a valid payload")
		}

		log.Printf("webhook received: %s/%s", strings.ToLower(payload.Body.Header.Headers["event_source"]), strings.ToLower(payload.Body.Header.Headers["event_type"]))
		return c.String(http.StatusOK, "OK")
	})

	e.DELETE("/webhook", func(c echo.Context) error {
		log.Printf("webhook received: ondelete/%s %s-%s", c.QueryParam("source"), c.QueryParam("namespace"), c.QueryParam("name"))
		return c.String(http.StatusOK, "OK")
	})

	log.Println("server listening at", addr)

	go func() {
		if err := e.Start(addr); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	runtime.SetupSignal(cancel)

	<-ctx.Done()

	defer cancel()

	log.Println("stopping server")
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
