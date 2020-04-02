package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/youngkin/todoshaleapps/src/internal/platform/constants"
	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

func (h handler) handlePut(w http.ResponseWriter, r *http.Request) {
	// parseRqst() logs parsing errors, no need to log again
	td, pathNodes, err := h.parseRqst(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(pathNodes) != 2 {
		httpStatus := http.StatusBadRequest
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: fmt.Sprintf("expecting resource path like /todos/{id}, got %+v", pathNodes),
		}).Error(constants.MalformedURL)
		w.WriteHeader(httpStatus)
		return
	}

	if pathNodes[1] != strconv.Itoa(td.ID) {
		httpStatus := http.StatusBadRequest
		h.logger.WithFields(log.Fields{
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: fmt.Sprintf("resource ID in url (%s) doesn't match resource ID in request body (%d)", pathNodes[1], td.ID),
		}).Error(constants.ToDoRqstErrorCode)
		w.WriteHeader(httpStatus)
		return
	}

	errCode, err := todo.UpdateToDo(h.db, td)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if errCode == constants.DBInvalidRequestCode {
			httpStatus = http.StatusBadRequest
		}
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(errCode)
		w.WriteHeader(httpStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
}
