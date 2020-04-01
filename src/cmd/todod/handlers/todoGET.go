package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
	"github.com/youngkin/todoshaleapps/src/internal/platform/constants"
	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

func (h handler) handleGet(w http.ResponseWriter, r *http.Request) {
	var httpStatus int
	// Expecting a URL.Path like '/todos' or '/todos/{id}'
	pathNodes, err := h.getURLPathNodes(r.URL.Path)
	if err != nil {
		httpStatus = http.StatusBadRequest
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)
		w.WriteHeader(httpStatus)
		return
	}

	var (
		payload   interface{}
		errReason constants.ErrCode
	)

	if len(pathNodes) == 1 {
		payload, errReason, err = h.handleGetToDoList(pathNodes[0])
	} else {
		payload, errReason, err = h.handleGetToDoItem(pathNodes[0], pathNodes[1:])
	}

	if err != nil {
		httpStatus = http.StatusInternalServerError
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errReason,
			constants.ErrorDetail: err.Error(),
			constants.HTTPStatus:  httpStatus,
		}).Error(constants.ToDoRqstError)
		if errReason == constants.MalformedURLErrorCode {
			// Malformed URL means that the request/URL was incorrectly specified.
			// This is a special case where http.StatusInternalServerError
			// isn't applicable.
			httpStatus = http.StatusBadRequest
		}
		w.WriteHeader(httpStatus)
		return
	}

	todoFound := true
	switch p := payload.(type) {
	case nil:
		todoFound = false
	case *todo.Item:
		if p == nil {
			todoFound = false
		}
	case *todo.List:
		if len(p.Items) == 0 {
			todoFound = false
		}
	default:
		httpStatus = http.StatusInternalServerError
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:  constants.ToDoTypeConversionErrorCode,
			constants.HTTPStatus: httpStatus,
		}).Error(constants.ToDoTypeConversionError)
		w.WriteHeader(httpStatus)
		return
	}

	if !todoFound {
		httpStatus = http.StatusNotFound
		h.logger.WithFields(log.Fields{
			constants.HTTPStatus: httpStatus,
		}).Error("ToDo not found")
		w.WriteHeader(httpStatus)
		return
	}

	marshPayload, err := json.Marshal(payload)
	if err != nil {
		httpStatus = http.StatusInternalServerError
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONMarshalingErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONMarshalingError)
		w.WriteHeader(httpStatus)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshPayload)
}

// handleGetToDoList will return the todo list, an error reason and error if there
// was a problem retrieving the todo, or a nil todo and a nil error if the todo was
// not found. The error reason will only be relevant when the error is non-nil.
func (h handler) handleGetToDoList(path string) (item interface{}, errReason constants.ErrCode, err error) {
	tds, err := todo.GetToDoList(h.db)
	if err != nil {
		return nil, constants.ToDoRqstErrorCode, errors.Annotate(err, "Error retrieving todos from DB")
	}

	h.logger.Debugf("handleGetToDoList() results: %+v", tds)

	for _, todo := range tds.Items {
		todo.SelfRef = "/" + path + "/" + strconv.Itoa(todo.ID)
	}

	return tds, constants.NoErrorCode, nil
}

// handleGetToDoItem will return the todo referenced by the provided resource path,
// an error reason and error if there was a problem retrieving the todo, or a nil todo and a nil
// error if the todo was not found. The error reason will only be relevant when the error
// is non-nil.
func (h handler) handleGetToDoItem(path string, pathNodes []string) (item interface{}, errReason constants.ErrCode, err error) {
	if len(pathNodes) > 1 {
		err := errors.Errorf(("expected 1 pathNode, got %d: path %s"), len(pathNodes), pathNodes)
		return nil, constants.MalformedURLErrorCode, err
	}

	id, err := strconv.Atoi(pathNodes[0])
	if err != nil {
		err := errors.Annotate(err, fmt.Sprintf("expected numeric pathNode, got %d", id))
		return nil, constants.MalformedURLErrorCode, err
	}

	td, err := todo.GetToDoItem(h.db, id)
	if err != nil {
		return nil, constants.ToDoRqstErrorCode, err
	}
	if td == nil {
		// client will deal with a nil (e.g., not found) todo
		return nil, constants.NoErrorCode, nil
	}

	h.logger.Debugf("GetToDoItem() results: %+v", td)

	td.SelfRef = "/" + path + "/" + strconv.Itoa(td.ID)

	return td, 0, nil
}
