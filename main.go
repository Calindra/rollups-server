package main

import (
	"fmt"
	"time"

	"github.com/calindra/rollups-server/src/model"
	"github.com/calindra/rollups-server/src/rollup"
	"github.com/calindra/rollups-server/src/sequencer"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const DefaultRollupsPort = 5004
const HttpTimeout = 10 * time.Second

func main() {

	db := sqlx.MustConnect("sqlite3", "file:memory1?mode=memory&cache=shared")
	//decoder := container.GetOutputDecoder()

	modelInstance := model.NewNonodoModel(nil, db)

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		ErrorMessage: "Request timed out",
		Timeout:      HttpTimeout,
	}))

	inputBoxSequencer := sequencer.NewInputBoxSequencer(modelInstance)

	rollup.Register(e, modelInstance, inputBoxSequencer)
	e.Start(fmt.Sprintf(":%s", DefaultRollupsPort))
}
