package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

func TestPUTToDo(t *testing.T) {
	client := &http.Client{}

	date := time.Date(2020, 4, 2, 13, 13, 0, 0, time.UTC)

	tcs := []TestCase{
		// {
		// 	testName:           "testUpdateToDoItemSuccess",
		// 	shouldPass:         true,
		// 	url:                "/todos/1",
		// 	expectedHTTPStatus: http.StatusOK,
		// 	expectedResourceID: "/todos/1",
		// 	postData:           `{"id":1,"note": "walk the dog","duedate":"2020-04-02T13:13:13Z","repeat": true,"completed": false}`,
		// 	todo: todo.Item{
		//		ID:        1,
		// 		Note:      "walk the dog",
		// 		DueDate:   now,
		// 		Repeat:    true,
		// 		Completed: false,
		// 	},
		// 	setupFunc:    todo.DBUpdateSetupHelper,
		// 	teardownFunc: todo.DBCallTeardownHelper,
		// },
		// {
		// 	testName:           "testUpdateDBError",
		// 	shouldPass:         false,
		// 	url:                "/todos/100",
		// 	expectedHTTPStatus: http.StatusInternalServerError,
		// 	updateResourceID:   "todos/100",
		// 	expectedResourceID: "",
		// 	postData:           "{\"id\":100,\"note\":\"walk the dog\",\"duedate\":\"2020-04-02T13:13:13Z\",\"repeat\":true,\"completed\":false}",
		// 	todo: todo.Item{
		// 		ID:        100,
		// 		Note:      "walk the dog",
		// 		DueDate:   date,
		// 		Repeat:    true,
		// 		Completed: false,
		// 	},
		// 	setupFunc:    todo.DBUpdateErrorSetupHelper,
		// 	teardownFunc: todo.DBCallTeardownHelper,
		// },
		{
			// ID in URL, '/todos/100', doesn't match ID in postData, '1' and todo '1'
			testName:           "testUpdateValidationError",
			shouldPass:         false,
			url:                "/todos/100",
			expectedHTTPStatus: http.StatusBadRequest,
			updateResourceID:   "todos/100",
			expectedResourceID: "",
			postData:           `{"id":1,"note":"walk the dog","duedate":"2020-04-02T13:13:13Z","repeat":true,"completed":false}`,
			todo: todo.Item{
				ID:        1,
				Note:      "walk the dog",
				DueDate:   date,
				Repeat:    true,
				Completed: false,
			},
			setupFunc:    todo.DBUpdateNoExpectationsSetupHelper,
			teardownFunc: todo.DBCallTeardownHelper,
		},
		{
			testName:           "testPUTInvalidURLMissingResourceID",
			shouldPass:         false,
			url:                "/todos",
			expectedHTTPStatus: http.StatusBadRequest,
			updateResourceID:   "todos/100",
			expectedResourceID: "",
			postData:           `{"id":1,"note":"walk the dog","duedate":"2020-04-02T13:13:13Z","repeat":true,"completed":false}`,
			todo: todo.Item{
				ID:        1,
				Note:      "walk the dog",
				DueDate:   date,
				Repeat:    true,
				Completed: false,
			},
			setupFunc:    todo.DBUpdateNoExpectationsSetupHelper,
			teardownFunc: todo.DBCallTeardownHelper,
		},
		{
			testName:           "testPUTNonNumericResourceID",
			shouldPass:         false,
			url:                "/todos/somebadnumber",
			expectedHTTPStatus: http.StatusBadRequest,
			updateResourceID:   "todos/100",
			expectedResourceID: "",
			postData:           `{"id":1,"note":"walk the dog","duedate":"2020-04-02T13:13:13Z","repeat":true,"completed":false}`,
			todo: todo.Item{
				ID:        1,
				Note:      "walk the dog",
				DueDate:   date,
				Repeat:    true,
				Completed: false,
			},
			setupFunc:    todo.DBUpdateNoExpectationsSetupHelper,
			teardownFunc: todo.DBCallTeardownHelper,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.testName, func(t *testing.T) {
			db, mock := tc.setupFunc(t, tc.todo)

			srvHandler, err := NewToDoHandler(db, logger)
			if err != nil {
				t.Fatalf("error '%s' was not expected when getting a todo handler", err)
			}

			testSrv := httptest.NewServer(http.HandlerFunc(srvHandler.ServeHTTP))
			defer testSrv.Close()

			url := testSrv.URL + tc.url
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(tc.postData)))
			if err != nil {
				t.Fatalf("an error '%s' was not expected creating HTTP request", err)
			}

			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("an error '%s' was not expected calling (client.Do()) todod server", err)
			}

			status := resp.StatusCode
			if status != tc.expectedHTTPStatus {
				t.Errorf("expected StatusCode = %d, got %d", tc.expectedHTTPStatus, status)
			}

			tc.teardownFunc(t, mock)
		})
	}
}
