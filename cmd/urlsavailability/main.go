package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/s-turchinskiy/urlsavailability/cmd/config"
	"github.com/s-turchinskiy/urlsavailability/internal/httpserver"
	"github.com/s-turchinskiy/urlsavailability/internal/repository/memcashed"
	"github.com/s-turchinskiy/urlsavailability/internal/service"
	"github.com/s-turchinskiy/urlsavailability/internal/utils/closerutil"
)

func main() {

	err := godotenv.Load("./cmd/urlsavailability/.env")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("Error loading .env file", "error", err.Error())
	}

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	closer := closerutil.New(cfg.ShutdownTimeout)

	var serv service.Servicer
	serv = service.New(memcashed.New(), cfg.URLTimeout, cfg.RateLimit, cfg.FileStoragePath)

	err = serv.LoadDataFromFile(ctx)
	if err != nil {
		log.Fatal(err)
	}

	httpHandlers := httpserver.NewHandlers(serv)
	httpServer := httpserver.NewHTTPServer(httpHandlers, cfg.Addr.String())

	closer.Add(httpServer.FuncShutdown())
	closer.Add(serv.SaveDataToFile)

	var wg sync.WaitGroup

	go func() {
		defer wg.Done()

		wg.Add(1)
		err = httpServer.Run()
		if err != nil {
			stop()
		}
	}()

	<-ctx.Done()
	err = closer.Shutdown()

	wg.Wait()

	if err != nil {
		log.Fatal(err)
	}
}
