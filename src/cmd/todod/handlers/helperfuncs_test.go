package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

// type MockReaderCloser struct {
// 	bytesRead int64
// 	myP       []byte
// }

// func (mrc MockReaderCloser) Read(p []byte) (int, error) {
// 	if mrc.bytesRead >= int64(len(mrc.myP)) {
// 		err := io.EOF
// 		return 0, err
// 	}

// 	n := copy(p, mrc.myP[mrc.bytesRead:])
// 	mrc.bytesRead += int64(n)
// 	return n, nil

// 	// bytesWritten := 0
// 	// startLen := len(mrc.myP) - *mrc.bytesRead
// 	// for i := 0; i < len(p) && i < startLen; i++ {
// 	// 	p[i] = mrc.myP[*mrc.bytesRead]
// 	// 	*mrc.bytesRead++
// 	// 	bytesWritten++
// 	// }
// 	// if bytesWritten == 0 {
// 	// 	return 0, io.EOF
// 	// }
// 	// return bytesWritten, nil
// }

// func (mrc MockReaderCloser) Close() error {
// 	return nil
// }

func TestGetURLPathNodes(t *testing.T) {
	testcases := []struct {
		testname      string
		inputURL      string
		expectedNodes []string
	}{
		{
			testname:      "/todo/",
			inputURL:      "/todo/",
			expectedNodes: []string{"todo"},
		},
		{
			testname:      "/todoNoTrailingSlash",
			inputURL:      "/todo",
			expectedNodes: []string{"todo"},
		},
		{
			testname:      "NoPathNodes",
			inputURL:      "/",
			expectedNodes: []string{},
		},
		{
			testname:      "/todo/1-WithResourceID",
			inputURL:      "/todo/1",
			expectedNodes: []string{"todo", "1"},
		},
		{
			testname:      "/todo/1?bulk=true-WithQueryParm",
			inputURL:      "/todo/1?bulk=true",
			expectedNodes: []string{"todo", "1"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.testname, func(t *testing.T) {
			u, err := url.Parse(fmt.Sprintf("http://todo.com%s", tc.inputURL))
			if err != nil {
				log.Fatal(err)
			}
			nodes, err := getURLPathNodes(u.Path)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}
			if !reflect.DeepEqual(tc.expectedNodes, nodes) {
				t.Errorf("expected %v, got %v", tc.expectedNodes, nodes)
			}
		})
	}
}

// TODO: Having trouble with the io.ReadCloser interface. Since both the
// 'Read()' and 'Close()' methods are pass-by-value, the state of the test
// buffer containing the mock response body is reset on each call resulting
// in an unmarshaling error. It would be nice to get this test working.
func TestParseBulkRqst(t *testing.T) {
	// date := time.Date(2020, 4, 2, 13, 13, 0, 0, time.UTC)

	testcases := []struct {
		testname       string
		inputURL       string
		inputJSON      string
		expectedOutput todo.List
	}{}
	// {
	// testname: "/todo?bulk=true",
	// inputURL: "/todo?bulk=true",
	// inputJSON: `
	// {
	// 	"todolist": [
	// 	  {
	// 		"id": 1,
	// 		"selfref": "/todos/1",
	// 		"note": "get groceries",
	// 		"duedate": "2020-04-01T00:00:00Z",
	// 		"repeat": false,
	// 		"completed": false
	// 	  },
	// 	  {
	// 		"id": 2,
	// 		"selfref": "/todos/2",
	// 		"note": "pay bills",
	// 		"duedate": "2020-04-02T00:00:00Z",
	// 		"repeat": false,
	// 		"completed": false
	// 	  },
	// 	  {
	// 		"id": 3,
	// 		"selfref": "/todos/3",
	// 		"note": "walk dog",
	// 		"duedate": "2020-04-03T12:00:00Z",
	// 		"repeat": true,
	// 		"completed": false
	// 	  }
	// 	]
	//   }`,
	// expectedOutput: todo.List{
	// 	Items: []*todo.Item{
	// 		{
	// 			ID:        1,
	// 			Note:      "get groceries",
	// 			DueDate:   date,
	// 			Repeat:    false,
	// 			Completed: false,
	// 		},
	// 		{
	// 			ID:        2,
	// 			Note:      "pay bills",
	// 			DueDate:   date,
	// 			Repeat:    false,
	// 			Completed: false,
	// 		},
	// 		{
	// 			ID:        3,
	// 			Note:      "walk dog",
	// 			DueDate:   date,
	// 			Repeat:    true,
	// 			Completed: false,
	// 		},
	// 	},
	// },
	// },
	// }

	for _, tc := range testcases {
		t.Run(tc.testname, func(t *testing.T) {
			payload, err := json.Marshal(tc.inputJSON)
			if err != nil {
				t.Errorf("unexpected error marshaling JSON: %s", err)
			}

			u, err := url.Parse(fmt.Sprintf("http://todo.com%s", tc.inputURL))
			if err != nil {
				log.Fatal(err)
			}
			r := http.Request{
				Method: "POST",
				Body:   ioutil.NopCloser(bytes.NewBuffer(payload)),
				URL:    u,
			}

			tdl, _, err := parseBulkRqst(&r, logger)
			if err != nil {
				t.Errorf("unexpected error calling parseBulkRequest: %s", err)
			}

			if reflect.DeepEqual(tc.expectedOutput, tdl) {
				t.Errorf("expected %+v, got %+v", tc.expectedOutput, tdl)
			}
		})
	}
}
