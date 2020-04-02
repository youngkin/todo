package handlers

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

type TestCase struct {
	testName           string
	shouldPass         bool
	url                string
	expectedHTTPStatus int
	updateResourceID   string
	expectedResourceID string
	postData           string
	todo               todo.Item
	setupFunc          func(*testing.T, todo.Item) (*sql.DB, sqlmock.Sqlmock)
	teardownFunc       func(*testing.T, sqlmock.Sqlmock)
}

func TestPOSTToDoItem(t *testing.T) {
	date := time.Date(2020, 4, 2, 13, 13, 0, 0, time.UTC)

	tcs := []TestCase{
		// {
		// 	testName:           "testInsertToDoItemSuccess",
		// 	shouldPass:         true,
		// 	url:                "/todos",
		// 	expectedHTTPStatus: http.StatusCreated,
		// 	expectedResourceID: "/todos/1",
		// 	postData:           `{"note": "walk the dog","duedate":"2020-04-02T13:13:13Z","repeat": true,"completed": false}`,
		// 	todo: todo.Item{
		// 		Note:      "walk the dog",
		// 		DueDate:   now,
		// 		Repeat:    true,
		// 		Completed: false,
		// 	},
		// 	setupFunc:    todo.DBInsertSetupHelper,
		// 	teardownFunc: todo.DBCallTeardownHelper,
		// },
		{
			// On insert the URL must not include a resource ID
			testName:           "testInsertToDoItemFailInvalidURL",
			shouldPass:         false,
			url:                "/todos/1",
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResourceID: "",
			postData:           `{"note": "walk the dog","duedate":"2020-04-02T13:13:13Z","repeat": true,"completed": false}`,
			todo: todo.Item{
				Note:      "walk the dog",
				DueDate:   date,
				Repeat:    true,
				Completed: false,
			},
			setupFunc:    todo.DBNoCallSetupHelper,
			teardownFunc: todo.DBCallTeardownHelper,
		},
		{
			// On insert the JSON body must not include todo ID
			testName:           "testInsertToDoItemFailInvalidJSON",
			shouldPass:         false,
			url:                "/todos",
			expectedHTTPStatus: http.StatusBadRequest,
			expectedResourceID: "",
			postData:           `{"id":1,"note": "walk the dog","duedate":"2020-04-02T13:13:13Z","repeat": true,"completed": false}`,
			todo: todo.Item{
				ID:        1,
				Note:      "walk the dog",
				DueDate:   date,
				Repeat:    true,
				Completed: false,
			},
			setupFunc:    todo.DBNoCallSetupHelper,
			teardownFunc: todo.DBCallTeardownHelper,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock := tc.setupFunc(t, tc.todo)

			srvHandler, err := NewToDoHandler(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a ToDo handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(srvHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL + tc.url
			resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(tc.postData)))
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling todod server", err)
			}
			defer resp.Body.Close()

			if tc.shouldPass {
				resourceURL := resp.Header.Get("Location")
				if string(resourceURL) != tc.expectedResourceID {
					t.Errorf("expected resource %s, got %s", tc.expectedResourceID, resourceURL)
				}
			}

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			tc.teardownFunc(t, mock)
		})
	}
}
