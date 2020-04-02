package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/todoshaleapps/src/internal/platform/constants"
	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

func (h handler) handlePost(w http.ResponseWriter, r *http.Request) {
	//
	// Get todo out of request body and validate
	//
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // error if todo has extra data
	todo := todo.Item{}
	err := d.Decode(&todo)
	if err != nil {
		httpStatus := http.StatusBadRequest
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONDecodingError)
		w.WriteHeader(httpStatus)
		return
	}
	if d.More() {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.ErrorDetail: err.Error(),
		}).Warn(constants.JSONDecodingError)
	}

	// Expecting t URL.Path '/todos'
	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		httpStatus := http.StatusBadRequest
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)
		w.WriteHeader(httpStatus)
		return
	}

	if todo.ID != 0 { // Item ID must *NOT* be populated (i.e., with a non-zero value) on an insert
		httpStatus := http.StatusBadRequest
		errMsg := fmt.Sprintf("expected Item.ID > 0, got Item.ID = %d", todo.ID)
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.InvalidInsertErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: errMsg,
		}).Error(constants.InvalidInsertError)
		w.WriteHeader(httpStatus)
		return
	}

	if len(pathNodes) != 1 {
		httpStatus := http.StatusBadRequest
		errMsg := fmt.Sprintf("expected '/todos', got %s", pathNodes)
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: errMsg,
		}).Error(constants.MalformedURL)
		w.WriteHeader(httpStatus)
		return
	}

	id, err := h.insertToDo(todo)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.DBUpSertErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.DBUpSertError)
		w.WriteHeader(httpStatus)
		return
	}

	w.Header().Add("Location", fmt.Sprintf("/todos/%d", id))
	w.WriteHeader(http.StatusCreated)
}

func (h handler) insertToDo(u todo.Item) (int64, error) {
	id, err := todo.InsertToDo(h.db, u)
	if err != nil {
		return -1, errors.Annotate(err, "error inserting todo")
	}
	return id, nil
}
