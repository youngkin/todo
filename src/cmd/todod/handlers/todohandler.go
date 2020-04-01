package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/todoshaleapps/src/internal/platform/constants"
	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

type handler struct {
	db     *sql.DB
	logger *log.Entry
}

// ServeHTTP handles the request
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logRqstRcvd(r)
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		fmt.Fprintf(w, "TODO - Implement!.")
		w.WriteHeader(http.StatusNotImplemented)
		// h.handlePost(w, r)
	case http.MethodPut:
		fmt.Fprintf(w, "TODO - Implement!.")
		w.WriteHeader(http.StatusNotImplemented)
		//		h.handlePut(w, r)
	case http.MethodDelete:
		fmt.Fprintf(w, "TODO - Implement!.")
		w.WriteHeader(http.StatusNotImplemented)
		// h.handleDelete(w, r)
	default:
		fmt.Fprintf(w, "Sorry, only GET, PUT, POST, and DELETE methods are supported.")
		w.WriteHeader(http.StatusNotImplemented)
	}

}

func (h handler) logRqstRcvd(r *http.Request) {
	h.logger.WithFields(log.Fields{
		constants.Method:     r.Method,
		constants.Path:       r.URL.Path,
		constants.RemoteAddr: r.RemoteAddr,
	}).Info("HTTP request received")
}

func (h handler) getURLPathNodes(path string) ([]string, error) {
	pathNodes := strings.Split(path, "/")

	if len(pathNodes) < 2 {
		return nil, errors.New(constants.ToDoRqstError)
	}

	// Strip off empty string that replaces the first '/' in '/todo'
	pathNodes = pathNodes[1:]

	// Strip off the empty string that replaces the second '/' in '/todo/'
	if pathNodes[len(pathNodes)-1] == "" {
		pathNodes = pathNodes[0 : len(pathNodes)-1]
	}

	return pathNodes, nil
}

func (h handler) parseRqst(r *http.Request) (todo.List, []string, error) {
	//
	// Get todo out of request body and validate
	//
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // error if todo sends extra data
	td := todo.List{}
	err := d.Decode(&td)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONDecodingError)

		return todo.List{}, nil, err
	}
	if d.More() {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.ErrorDetail: fmt.Sprintf("Additional JSON after ToDo data: %v", td),
		}).Warn(constants.JSONDecodingError)
	}

	// Expecting a URL.Path like '/todo/{id}'
	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)

		return todo.List{}, nil, err
	}

	return td, pathNodes, nil
}

// NewToDoHandler returns a *http.Handler configured with a database connection
func NewToDoHandler(db *sql.DB, logger *log.Entry) (http.Handler, error) {
	if db == nil {
		return nil, errors.New("non-nil sql.DB connection required")
	}
	if logger == nil {
		return nil, errors.New("non-nil log.Entry  required")
	}

	return handler{db: db, logger: logger}, nil
}
