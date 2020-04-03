package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

const nilToDoID = 0

// ServeHTTP handles the request
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logRqstRcvd(r, h.logger)
	bulk := r.URL.Query().Get("bulk")
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		if len(bulk) > 0 {
			h.handleBulkPost(w, r)
			return
		}
		//
		// Get todo out of request body and validate
		//
		td, pathNodes, err := parseRqst(r, h.logger)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h.handlePost(w, r, td, pathNodes)
	case http.MethodPut:
		h.handlePut(w, r)
	case http.MethodDelete:
		h.handleDelete(w, r)
	default:
		fmt.Fprintf(w, "Sorry, only GET, PUT, POST, and DELETE methods are supported.")
		w.WriteHeader(http.StatusNotImplemented)
	}

}

func (h handler) handleGet(w http.ResponseWriter, r *http.Request) {
	var httpStatus int
	// Expecting a URL.Path like '/todos' or '/todos/{id}'
	pathNodes, err := getURLPathNodes(r.URL.Path)
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

	for _, td := range tds.Items {
		td.SelfRef = "/" + path + "/" + strconv.FormatInt(td.ID, 10)
	}

	return &tds, constants.NoErrorCode, nil
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
		// caller will deal with a nil (e.g., not found) todo
		return nil, constants.NoErrorCode, nil
	}

	h.logger.Debugf("GetToDoItem() results: %+v", td)

	td.SelfRef = "/" + path + "/" + strconv.FormatInt(td.ID, 10)

	return td, 0, nil
}

func (h handler) handlePost(w http.ResponseWriter, r *http.Request, td todo.Item, pathNodes []string) {
	// An expected request will not include Item.ID and the resulting unMarshaled Item.ID
	// will take it's zero-value of 0. In Postres (and MySQL) the 'SERIAL' datatype's first
	// value will be '1' so '0', or 'nilToDoID', is a valid indication of an unset Item.ID.
	if td.ID != nilToDoID {
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

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Location", fmt.Sprintf("/todos/%d", id))
}

func (h handler) handleBulkPost(w http.ResponseWriter, r *http.Request) {
	tdl, pathNodes, err := parseBulkRqst(r, h.logger)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	numRqsts := 0
	for _, td := range tdl.Items {
		td := td
		rqst := insertTodoRequest{
			r:         r,
			td:        *td,
			pathNodes: pathNodes,
			respChan:  insertToDoRespChan,
		}
		InsertToDoRqstChan <- rqst
		numRqsts++
	}

	h.logger.WithFields(log.Fields{
		constants.Method: http.MethodPost,
	}).Debugf("handleBulkPost, launched %d insert requests", numRqsts)

	var responses = []insertTodoResponse{}
	for i := 0; i < numRqsts; i++ {
		resp := <-insertToDoRespChan
		h.logger.WithFields(log.Fields{
			constants.Method:        http.MethodPost,
			constants.MessageDetail: fmt.Sprintf("Response: %+v", resp),
		}).Debugf("handleBulkPost received response")

		if resp.HTTPStatus == http.StatusCreated {
			resp.Item.SelfRef = "/" + pathNodes[0] + "/" + strconv.FormatInt(resp.Item.ID, 10)
		}

		responses = append(responses, resp)
	}

	respOut := insertTodoResponses{
		Responses: responses,
	}

	marshResp, err := json.Marshal(respOut)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONMarshalingErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONMarshalingError)
		w.WriteHeader(httpStatus)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(marshResp)
}

func (h handler) handlePostItem(r *http.Request, td todo.Item, pathNodes []string, respChan chan insertTodoResponse) {
	h.logger.WithFields(log.Fields{
		constants.Method:        http.MethodPost,
		constants.MessageDetail: fmt.Sprintf("Item: %+v", td),
	}).Debugf("handlePostItem entry")
	// An expected request will not include Item.ID and the resulting unMarshaled Item.ID
	// will take it's zero-value of 0. In Postres (and MySQL) the 'SERIAL' datatype's first
	// value will be '1' so '0', or 'nilToDoID', is a valid indication of an unset Item.ID.
	if td.ID != nilToDoID {
		httpStatus := http.StatusBadRequest
		errMsg := fmt.Sprintf("expected Item.ID > 0, got Item.ID = %d", td.ID)
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.InvalidInsertErrorCode,
			constants.HTTPStatus:  httpStatus,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: errMsg,
		}).Error(constants.InvalidInsertError)
		resp := insertTodoResponse{
			Item:       td,
			HTTPStatus: httpStatus,
			Err:        errors.Errorf(errMsg),
		}
		respChan <- resp
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
		resp := insertTodoResponse{
			Item:       td,
			HTTPStatus: httpStatus,
			Err:        errors.Errorf(errMsg),
		}
		respChan <- resp
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
		resp := insertTodoResponse{
			Item:       td,
			HTTPStatus: httpStatus,
			Err:        errors.Annotate(err, "call to insertToDo() failed"),
		}
		respChan <- resp
	}

	td.ID = id
	resp := insertTodoResponse{
		Item:       td,
		HTTPStatus: http.StatusCreated,
		Err:        nil,
	}
	respChan <- resp
	h.logger.WithFields(log.Fields{
		constants.Method:        http.MethodPost,
		constants.MessageDetail: fmt.Sprintf("Response: %+v", resp),
	}).Debugf("handlePostItem exit")
}

func (h handler) insertToDo(u todo.Item) (int64, error) {
	id, err := todo.InsertToDo(h.db, u)
	if err != nil {
		return -1, errors.Annotate(err, "error inserting todo")
	}
	return id, nil
}

func (h handler) handlePut(w http.ResponseWriter, r *http.Request) {
	// parseRqst() logs parsing errors, no need to log again
	td, pathNodes, err := parseRqst(r, h.logger)
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

	if pathNodes[1] != strconv.FormatInt(td.ID, 10) {
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

func (h handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	pathNodes, err := getURLPathNodes(r.URL.Path)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(pathNodes) != 2 {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: fmt.Sprintf("expecting resource path like /todos/{id}, got %+v", pathNodes),
		}).Error(constants.MalformedURL)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uid, err := strconv.Atoi(pathNodes[1])
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: fmt.Sprintf("Invalid resource ID, must be int, got %v", pathNodes[1]),
		}).Error(constants.MalformedURL)

		w.WriteHeader(http.StatusBadRequest)
		return
	}
	errCode, err := todo.DeleteToDo(h.db, uid)
	if err != nil {
		h.logger.WithFields(log.Fields{
			constants.ErrorCode:   errCode,
			constants.HTTPStatus:  http.StatusInternalServerError,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(errCode)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func logRqstRcvd(r *http.Request, logger *log.Entry) {
	logger.WithFields(log.Fields{
		constants.Method:     r.Method,
		constants.Path:       r.URL.Path,
		constants.RemoteAddr: r.RemoteAddr,
	}).Info("HTTP request received")
}

func getURLPathNodes(path string) ([]string, error) {
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

func parseRqst(r *http.Request, logger *log.Entry) (todo.Item, []string, error) {
	// Expecting a URL.Path like '/todos/' or '/todos?bulk=true'
	pathNodes, err := getURLPathNodes(r.URL.Path)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)

		return todo.Item{}, nil, errors.Annotate(err, "error occurred while extracting URL path nodes")
	}

	//
	// Get todo out of request body and validate
	//
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // error if todo sends extra data
	td := todo.Item{}
	err = d.Decode(&td)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONDecodingError)

		return todo.Item{}, nil, errors.Annotate(err, "error occurred while unmarshaling request body")
	}
	if d.More() {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.ErrorDetail: fmt.Sprintf("Additional JSON after ToDo data: %v", td),
		}).Warn(constants.JSONDecodingError)
	}

	return td, pathNodes, nil
}

func parseBulkRqst(r *http.Request, logger *log.Entry) (todo.List, []string, error) {
	// Expecting a URL.Path like '/todos/' or '/todos?bulk=true'
	pathNodes, err := getURLPathNodes(r.URL.Path)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.MalformedURLErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.Path:        r.URL.Path,
			constants.ErrorDetail: err,
		}).Error(constants.MalformedURL)

		return todo.List{}, nil, errors.Annotate(err, "error occurred while extracting URL path nodes")
	}

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields() // error if todo sends extra data
	tdl := todo.List{}
	err = d.Decode(&tdl)
	if err != nil {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.HTTPStatus:  http.StatusBadRequest,
			constants.ErrorDetail: err.Error(),
		}).Error(constants.JSONDecodingError)

		return todo.List{}, nil, errors.Annotate(err, "error occurred while unmarshaling request body")
	}
	if d.More() {
		logger.WithFields(log.Fields{
			constants.ErrorCode:   constants.JSONDecodingErrorCode,
			constants.ErrorDetail: fmt.Sprintf("Additional JSON after ToDo data: %v", tdl),
		}).Warn(constants.JSONDecodingError)
	}

	return tdl, pathNodes, nil
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

type insertTodoResponse struct {
	Item       todo.Item `json:"item"`
	HTTPStatus int       `json:"httpStatus"`
	Err        error     `json:"error"`
}

type insertTodoResponses struct {
	Responses []insertTodoResponse `json:"responses"`
}

type insertTodoRequest struct {
	r         *http.Request
	td        todo.Item
	pathNodes []string
	respChan  chan insertTodoResponse
}

var (
	maxSimToDoInserts = 10
	// ToDoPostDoneChan is used during shutdown to stop daemon goroutines gracefully
	ToDoPostDoneChan = make(chan interface{})
	// InsertToDoRqstChan is to send insert To Do Item requests for processing
	InsertToDoRqstChan = make(chan insertTodoRequest, maxSimToDoInserts)
	insertToDoRespChan = make(chan insertTodoResponse)
)

// PostRequestLauncher listens on its 'rqstChan' for insert To Do Item requests and
// will launch each request into a goroutine for processing. It monitors its 'done'
// channel to detect when it should exit.
func PostRequestLauncher(h interface{}, done chan interface{}, rqstChan chan insertTodoRequest, logger *log.Entry) {
	var hndlr handler

	switch h.(type) {
	case handler:
		hndlr = h.(handler)
	default:
		logger.Fatalf("PostRequestLauncher provided an 'h' parameter that is not of type 'handler'")
	}
	logger.Info("PostRequestLauncher starting...")
	for {
		select {
		case rqst := <-rqstChan:
			go hndlr.handlePostItem(rqst.r, rqst.td, rqst.pathNodes, rqst.respChan)
		case <-done:
			logger.Info("PostRequestLauncher exiting...")
			return
		default:
			continue
		}
	}
}
