package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/calindra/rollups-server/src/container"
	"github.com/calindra/rollups-server/src/devnet"
	"github.com/calindra/rollups-server/src/model"
	"github.com/calindra/rollups-server/src/rollup"
	"github.com/calindra/rollups-server/src/sequencer"
	"github.com/calindra/rollups-server/src/sequencer/inputter"
	"github.com/calindra/rollups-server/src/supervisor"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

const DefaultRollupsPort = 5004
const HttpTimeout = 10 * time.Second

var startupMessage = `
Http Rollups for development started at http://localhost:5004
Press Ctrl+C to stop the server
`

func main() {
	startTime := time.Now()
	var w supervisor.SupervisorWorker
	db := sqlx.MustConnect("sqlite3", "file:memory1?mode=memory&cache=shared")
	container := container.NewContainer(*db)
	decoder := container.GetOutputDecoder()

	modelInstance := model.NewAppModel(decoder, db)

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339} ${remote_ip} ${method} ${uri} ${status} ${latency_human}` + "\n",
	}))
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		ErrorMessage: "Request timed out",
		Timeout:      HttpTimeout,
	}))

	inputBoxSequencer := sequencer.NewInputBoxSequencer(modelInstance)

	w.Workers = append(w.Workers, devnet.AnvilWorker{
		Address: devnet.AnvilDefaultAddress,
		Port:    devnet.AnvilDefaultPort,
		Verbose: false,
	})

	w.Workers = append(w.Workers, inputter.InputterWorker{
		Model:              modelInstance,
		Provider:           fmt.Sprintf("ws://%s:%v", devnet.AnvilDefaultAddress, devnet.AnvilDefaultPort),
		InputBoxAddress:    common.HexToAddress(devnet.InputBoxAddress),
		InputBoxBlock:      0,
		ApplicationAddress: common.HexToAddress(devnet.ApplicationAddress),
	})

	rollup.Register(e, modelInstance, inputBoxSequencer)

	w.Workers = append(w.Workers, supervisor.HttpWorker{
		Address: "127.0.0.1:5004",
		Handler: e,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ready := make(chan struct{}, 1)
	go func() {
		select {
		case <-ready:
			fmt.Println(startupMessage)
			slog.Info("http server: ready", "after", time.Since(startTime))
		case <-ctx.Done():
		}
	}()
	w.Start(ctx, ready)

}
