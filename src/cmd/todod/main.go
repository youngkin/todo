package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/todoshaleapps/src/internal/logging"
	"github.com/youngkin/todoshaleapps/src/internal/platform/constants"
)

// TODO:
//	1. 	Get Travis CI working in github
//	2. 	Super simple, hello world, web service in google cloud platform and tested
//	3.	Setup DB in GCP with tables and sample data
//	4.	Get GET '/todo' and '/todo/{id}' endpoints, unit tests, and associated DB test data working
//	5.	Finish PUT/POST/DELETE with tests
//	6.	'curl' examples in README

func main() {
	// configFileName := flag.String("configFile",
	// 	"/opt/mockvideo/accountd/config/config",
	// 	"specifies the location of the accountd service configuration")
	// secretsDir := flag.String("secretsDir",
	// 	"/opt/mockvideo/accountd/secrets",
	// 	"specifies the location of the accountd secrets")

	logLevel := flag.Int("loglevel", 4,
		"specifies the logging level, 4(INFO) is the default. Levels run from 0 (PANIC) to 6 (TRACE)")
	port := flag.Int("port", 8080, "specifies this service's listenting port")
	flag.Parse()

	// 'logger' comes set with a default log level. This will be used if there's a problem
	// with the provided log level.
	logger := logging.GetLogger().WithField(constants.Application, "ToDo")
	log.SetLevel(log.Level(*logLevel))

	//
	// Setup DB connection
	//
	// connStr, err := getDBConnectionStr(configs, secrets)
	// if err != nil {
	// 	logger.WithFields(log.Fields{
	// 		constants.ErrorCode:   constants.UnableToGetDBConnStrErrorCode,
	// 		constants.ErrorDetail: err.Error(),
	// 	}).Fatal(constants.UnableToGetDBConnStr)
	// }

	// db, err := sql.Open("mysql", connStr)
	// if err != nil {
	// 	logger.WithFields(log.Fields{
	// 		constants.ErrorCode:   constants.UnableToOpenDBConnErrorCode,
	// 		constants.ErrorDetail: err.Error(),
	// 	}).Fatal(constants.UnableToOpenDBConn)
	// }
	// defer db.Close()

	//
	// Setup endpoints and start service
	//
	// usersHandler, err := users.NewUserHandler(db, logger)
	// if err != nil {
	// 	logger.WithFields(log.Fields{
	// 		constants.ErrorCode:   constants.UnableToCreateHTTPHandlerErrorCode,
	// 		constants.ErrorDetail: err.Error(),
	// 	}).Fatal(constants.UnableToCreateHTTPHandler)
	// }

	// healthHandler := http.HandlerFunc(handlers.HealthFunc)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.WithFields(log.Fields{
			constants.ServiceName: "health",
		}).Info("handling request")
		w.Write([]byte("I'm healthy!\n"))
	})

	// port, ok := configs["port"]
	// if !ok {
	// 	logger.Info("port configuration unavailable (configs[port]), defaulting to 5000")
	// 	port = "5000"
	// }
	// port = ":" + port

	addr := ":" + strconv.Itoa(*port)
	s := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	go func() {
		logger.WithFields(log.Fields{
			constants.Port:     addr,
			constants.LogLevel: log.GetLevel().String(),
		}).Info("todod service starting")

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	handleTermSignal(s, logger, 10)
}

//
// Helper funcs
//

// handleTermSignal provides a mechanism to catch SIGTERMs and gracefully
// shutdown the service.
func handleTermSignal(s *http.Server, logger *log.Entry, timeout int) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	logger.Infof("Server shutting down with timeout: %d", timeout)

	if err := s.Shutdown(ctx); err != nil {
		logger.Warnf("Server shutting down with error: %s", err)
	} else {
		logger.Info("Server stopped")
	}

}
