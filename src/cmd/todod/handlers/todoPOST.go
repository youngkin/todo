package handlers

import (
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
	td, pathNodes, err := h.parseRqst(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// An expected request will not include Item.ID and the resulting unMarshaled Item.ID
	// will take it's zero-value of 0. In Postres (and MySQL) the 'SERIAL' datatype's first
	// value will be '1' so '0' is a valid indication of an unset Item.ID.
	if td.ID != 0 {
		httpStatus := http.StatusBadRequest
		errMsg := fmt.Sprintf("expected Item.ID > 0, got Item.ID = %d", td.ID)
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

	id, err := h.insertToDo(td)
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
