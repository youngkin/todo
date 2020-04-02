// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package todo

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// AnyTime is matcher for time.Time SQL statement arguments
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a *AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

// DBCallSetupHelper encapsulates common code needed to setup mock DB access to todo data
func DBCallSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, List) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "note", "duedate", "repeat", "completed"}).
		AddRow(1, "Get groceries", now, false, false).
		AddRow(2, "Walk Dog", now, true, false)

	mock.ExpectQuery(getAllToDosQuery).
		WillReturnRows(rows)

	expected := List{
		Items: []*Item{
			{
				ID:        1,
				SelfRef:   "/todos/1",
				Note:      "Get groceries",
				DueDate:   now,
				Repeat:    false,
				Completed: false,
			},
			{
				ID:        2,
				SelfRef:   "/todos/2",
				Note:      "Walk Dog",
				DueDate:   now,
				Repeat:    true,
				Completed: false,
			},
		},
	}

	return db, mock, expected
}

// DBCallTeardownHelper encapsulates common code needed to finalize processing of mock DB access to todo data
func DBCallTeardownHelper(t *testing.T, mock sqlmock.Sqlmock) {
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// DBCallQueryErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallQueryErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, List) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery(getAllToDosQuery).
		WillReturnError(fmt.Errorf("some error"))

	return db, mock, List{}
}

// DBCallRowScanErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallRowScanErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, List) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"badRow"}).
		AddRow(-1)

	mock.ExpectQuery(getAllToDosQuery).
		WillReturnRows(rows)

	return db, mock, List{}
}

// GetItemSetupHelper encapsulates common code needed to setup mock DB access a single todo item's data
func GetItemSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Item) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "note", "duedate", "repeat", "completed"}).
		AddRow(1, "Get groceries", now, false, false)

	mock.ExpectQuery(getAllToDosQuery).
		WillReturnRows(rows)

	expected := Item{
		ID:        1,
		SelfRef:   "/todos/1",
		Note:      "Get groceries",
		DueDate:   now,
		Repeat:    false,
		Completed: false,
	}

	return db, mock, &expected
}

// DBCallNoExpectationsSetupHelper encapsulates common coded needed to when no expectations are present
func DBCallNoExpectationsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Item) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	return db, mock, nil
}

// DBUpdateNoExpectationsSetupHelper encapsulates common coded needed to when no expectations are present
func DBUpdateNoExpectationsSetupHelper(t *testing.T, td Item) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	return db, mock
}

// DBInsertSetupHelper encapsulates the common code needed to setup a mock To Do Item insert
func DBInsertSetupHelper(t *testing.T, td Item) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)

	mock.ExpectQuery(insertToDoStmt).
		WithArgs(td.Note, AnyTime{}, td.Repeat, td.Completed).
		WillReturnRows(rows)

	return db, mock
}

// DBUpdateSetupHelper encapsulates the common code needed to setup a mock Item update
func DBUpdateSetupHelper(t *testing.T, td Item) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	// mock.ExpectExec("UPDATE todo SET (.+) WHERE (.+)").WithArgs(TODO ADD).
	// 	WillReturnResult(sqlmock.NewResult(0, 1)) // no insert ID, 1 row affected
	mock.ExpectExec("UPDATE todo").WithArgs(sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg).
		WillReturnResult(sqlmock.NewResult(0, 1)) // no insert ID, 1 row affected
	return db, mock
}

// DBUpdateErrorSetupHelper encapsulates the common code needed to setup a mock Item update error
func DBUpdateErrorSetupHelper(t *testing.T, td Item) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectExec("UPDATE todo SET (.+) WHERE (.+)").WithArgs(sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg).WillReturnError(sql.ErrConnDone)
	return db, mock
}

// DBNoCallSetupHelper encapsulates the common code needed to mock an error upstream from an actual DB call
func DBNoCallSetupHelper(t *testing.T, u Item) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	return db, mock
}

// DBDeleteSetupHelper encapsulates the common code needed to setup a mock Item delete
func DBDeleteSetupHelper(t *testing.T, td Item) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectExec(deleteToDoStmt).WithArgs(td.ID)

	return db, mock
}
