package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"photoapi/internal/config"
	httpcontroller "photoapi/internal/controllers/http-controller"
	photoapi "photoapi/internal/photo-api"
	"photoapi/internal/photo-api/storage/sqlite"
)

func main() {
	err := run()
	if err != nil {
		log.Printf("run app: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// load .env file
	godotenv.Load()

	// Collecting prerequisites.
	gin.SetMode(gin.ReleaseMode)
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	if err != nil {
		return errors.Wrap(err, "initialising logger")
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Creating storage realisation via SQLite.
	dbConfig := config.NewDBConfig()
	db, err := sql.Open("sqlite3", dbConfig.Host)
	defer db.Close()
	storage := sqlite.New(db, dbConfig)

	// Applying the last version of storage schema.
	err = storage.MigrateUp(ctx)
	if err != nil {
		return errors.Wrap(err, "migrating storage up")
	}

	// Creating photo api service from collected dependencies.
	service := photoapi.New(storage)

	// Creating controllers.
	router := gin.Default()
	httpHandler := httpcontroller.New(router, service, config.NewHTTPConfig())

	// Starting contollers.
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := httpHandler.Start()
		if err != nil {
			logger.Error(fmt.Sprintf("http handler died\nError: %v", err))
		}
	}()

	wg.Wait()
	return errors.New("All handlers died")
}
