package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/todoshaleapps/src/internal/logging"
	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

type Test struct {
	testName           string
	url                string
	shouldPass         bool
	setupFunc          func(*testing.T) (*sql.DB, sqlmock.Sqlmock, todo.List)
	teardownFunc       func(*testing.T, sqlmock.Sqlmock)
	expectedHTTPStatus int
}

// ItemTest differs from 'Tests' in the setupFunc function signature returns a *todo.Item vs. todo.List
type ItemTest struct {
	testName           string
	url                string
	shouldPass         bool
	setupFunc          func(*testing.T) (*sql.DB, sqlmock.Sqlmock, *todo.Item)
	teardownFunc       func(*testing.T, sqlmock.Sqlmock)
	expectedHTTPStatus int
}

// logger is used to control code-under-test logging behavior
var logger *log.Entry

func init() {
	logger = logging.GetLogger()
	// Uncomment for more verbose logging
	// logger.Logger.SetLevel(log.DebugLevel)
	// Suppress all application logging
	logger.Logger.SetLevel(log.PanicLevel)
	// Uncomment for non-tty logging
	// log.SetFormatter(&log.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	//  })
}

func TestGetAllItems(t *testing.T) {
	tcs := []Test{
		{
			testName:           "testGetToDoListSuccess",
			url:                "/todos",
			shouldPass:         true,
			setupFunc:          todo.DBCallSetupHelper,
			teardownFunc:       todo.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			testName:           "testToDoListSuccessTrailingSlash",
			url:                "/todos/",
			shouldPass:         true,
			setupFunc:          todo.DBCallSetupHelper,
			teardownFunc:       todo.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			testName:           "testGetToDoListQueryFailure",
			url:                "/todos",
			shouldPass:         false,
			setupFunc:          todo.DBCallQueryErrorSetupHelper,
			teardownFunc:       todo.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusInternalServerError,
		},
		{
			testName:           "testGetToDoListRowScanFailure",
			url:                "/todos",
			shouldPass:         false,
			setupFunc:          todo.DBCallRowScanErrorSetupHelper,
			teardownFunc:       todo.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			// populate Item.SelfRef from Item.ID
			for _, todo := range expected.Items {
				todo.SelfRef = "/todos/" + strconv.FormatInt(todo.ID, 10)
			}

			todoHandler, err := NewToDoHandler(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a todo handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(todoHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL + tc.url
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling todod server", err)
			}
			defer resp.Body.Close()

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			if tc.shouldPass {
				actual, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("an error '%s' was not expected reading response body", err)
				}

				mExpected, err := json.Marshal(expected)
				if err != nil {
					t.Fatalf("an error '%s' was not expected Marshaling %+v", err, expected)
				}

				if bytes.Compare(mExpected, actual) != 0 {
					t.Errorf("expected %+v, got %+v", string(mExpected), string(actual))
				}
			}

			// we make sure that all post-conditions were met
			tc.teardownFunc(t, mock)
		})
	}
}

func TestGetOneItem(t *testing.T) {
	tcs := []ItemTest{
		{
			testName:           "testGetItemSuccess",
			url:                "/todos/1",
			shouldPass:         true,
			setupFunc:          todo.GetItemSetupHelper,
			teardownFunc:       todo.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusOK,
		},
		{
			testName:           "testGetItemURLTooLong",
			url:                "/todos/1/extraNode",
			shouldPass:         false,
			setupFunc:          todo.DBCallNoExpectationsSetupHelper,
			teardownFunc:       todo.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			testName:           "testGetItemURLNonNumericID",
			url:                "/todos/notanumber",
			shouldPass:         false,
			setupFunc:          todo.DBCallNoExpectationsSetupHelper,
			teardownFunc:       todo.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			testName:           "testGetItemErrNoRow",
			url:                "/todos/notanumber",
			shouldPass:         false,
			setupFunc:          todo.DBCallNoExpectationsSetupHelper,
			teardownFunc:       todo.DBCallTeardownHelper,
			expectedHTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock, expected := tc.setupFunc(t)
			defer db.Close()

			if expected != nil {
				expected.SelfRef = tc.url
			}

			todoHandler, err := NewToDoHandler(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a todo handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(todoHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL + tc.url
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling todod server", err)
			}
			defer resp.Body.Close()

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			if tc.shouldPass {
				actual, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("an error '%s' was not expected reading response body", err)
				}

				mExpected, err := json.Marshal(expected)
				if err != nil {
					t.Fatalf("an error '%s' was not expected Marshaling %+v", err, expected)
				}

				if bytes.Compare(mExpected, actual) != 0 {
					t.Errorf("expected %+v, got %+v", string(mExpected), string(actual))
				}
			}

			// we make sure that all post-conditions were met
			tc.teardownFunc(t, mock)
		})
	}
}
