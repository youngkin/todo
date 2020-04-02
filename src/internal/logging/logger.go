package logging

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/todoshaleapps/src/internal/platform/constants"
)

var logger *log.Entry

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Only log the INFO severity or above.
	log.SetLevel(log.InfoLevel)
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}

	logger = log.WithFields(log.Fields{
		constants.HostName: hostName,
	})

}

// GetLogger gets the common logger
func GetLogger() *log.Entry {
	return logger
}
