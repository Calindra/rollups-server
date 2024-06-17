package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const DefaultHttpPort = 8080
const HttpTimeout = 10 * time.Second

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var users = []User{
	{ID: 1, Name: "Alice"},
	{ID: 2, Name: "Bob"},
}

func main() {
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		ErrorMessage: "Request timed out",
		Timeout:      HttpTimeout,
	}))

	e.GET("/users", getUsers)

	e.Start(":8080")
}

func getUsers(c echo.Context) error {
	return c.JSON(http.StatusOK, users)
}
