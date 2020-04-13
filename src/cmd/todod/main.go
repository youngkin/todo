package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/todoshaleapps/src/cmd/todod/handlers"
	"github.com/youngkin/todoshaleapps/src/internal/logging"
	"github.com/youngkin/todoshaleapps/src/internal/platform/constants"
)

func main() {
	logger := logging.GetLogger().WithField(constants.Application, "ToDo")

	dbportStr, ok := os.LookupEnv("POSTGRES_SERVICE_PORT")
	if !ok {
		logger.Warn("$POSTGRES_SERVICE_PORT not found, defaulting to 5432")
		dbportStr = "5432"
	}
	dfltDbport, err := strconv.Atoi(dbportStr)
	if err != nil {
		logger.Warn("error converting POSTGRES_SERVICE_PORT to int, defaulting to 5432", dbportStr)
		dfltDbport = 5432
	}
	dfltDbhost, ok := os.LookupEnv("POSTGRES_SERVICE_HOST")
	if !ok {
		logger.Warn("$POSTGRES_SERVICE_HOST not found, defaulting to localhost")
		dfltDbhost = "localhost"
	}

	logLevel := flag.Int("loglevel", 4,
		"specifies the logging level, 4(INFO) is the default. Levels run from 0 (PANIC) to 6 (TRACE)")
	port := flag.Int("port", 8080, "specifies this service's listening port")
	// Normally, info like this should NEVER come from the command line.
	dbPort := flag.Int("dbport", dfltDbport, "specifies the database's connection port")
	dbHost := flag.String("dbhost", dfltDbhost, "specifies the hostname or address of the database server")
	dbUser := flag.String("dbuser", "todo", "DB user's login ID")
	password := flag.String("passwd", "todo123", "DB user's password")
	dbName := flag.String("dbname", "todo", "application's db name")

	flag.Parse()

	// 'logger' comes set with a default log level. This will be used if there's a problem
	// with the provided log level.
	log.SetLevel(log.Level(*logLevel))

	//
	// Setup DB connection
	//
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", *dbHost, *dbPort, *dbUser, *password, *dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.UnableToOpenDBConnErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Fatal(constants.UnableToOpenDBConn)
	}
	err = db.Ping()
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.UnableToOpenDBConnErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Fatal(constants.UnableToOpenDBConn + ": database unreachable")
	}
	defer db.Close()

	//
	// Setup endpoints and start service
	//
	todoHandler, err := handlers.NewToDoHandler(db, logger)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.UnableToCreateHTTPHandlerErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Fatal(constants.UnableToCreateHTTPHandler)
	}

	mux := http.NewServeMux()
	mux.Handle("/todos", todoHandler) // Adding this route is necessary to support query parms like /todos?bulk=true
	mux.Handle("/todos/", todoHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(log.Fields{
			constants.ServiceName: "health",
		}).Info("handling request")
		w.Write([]byte("I'm healthy!\n"))
	})

	addr := ":" + strconv.Itoa(*port)
	s := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	go handlers.PostRequestLauncher(todoHandler, handlers.ToDoPostDoneChan, handlers.InsertToDoRqstChan, logger)

	go func() {
		logger.WithFields(log.Fields{
			constants.Port:     addr,
			constants.LogLevel: log.GetLevel().String(),
			constants.DBHost:   *dbHost,
			constants.DBPort:   *dbPort,
		}).Info("todod service starting")

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			close(handlers.ToDoPostDoneChan)
			logger.Fatal(err)
		}
	}()

	handleTermSignal(s, logger, 10)
}

// handleTermSignal provides a mechanism to catch SIGTERMs and gracefully
// shutdown the service.
func handleTermSignal(s *http.Server, logger *log.Entry, timeout int) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	close(handlers.ToDoPostDoneChan)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	logger.Infof("Server shutting down with timeout: %d", timeout)

	if err := s.Shutdown(ctx); err != nil {
		logger.Warnf("Server shutting down with error: %s", err)
	} else {
		logger.Info("Server stopped")
	}
}
