// Copyright (c) 2020 Richard Youngkin. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package handlers

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/youngkin/todoshaleapps/src/internal/todo"
)

// DBCallSetupHelper encapsulates common code needed to setup mock DB access to todo data
func DBCallSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, todo.List) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "note", "duedate", "repeat", "completed"}).
		AddRow(1, "Get groceries", now, false, false).
		AddRow(2, "Walk Dog", now, true, false)

	mock.ExpectQuery("SELECT id, note, duedate, repeat, completed FROM todo").
		WillReturnRows(rows)

	expected := todo.List{
		Items: []*todo.Item{
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
func DBCallQueryErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, todo.List) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	mock.ExpectQuery("SELECT id, note, duedate, repeat, completed FROM todo").
		WillReturnError(fmt.Errorf("some error"))

	return db, mock, todo.List{}
}

// DBCallRowScanErrorSetupHelper encapsulates common coded needed to mock DB query failures
func DBCallRowScanErrorSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, todo.List) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	rows := sqlmock.NewRows([]string{"badRow"}).
		AddRow(-1)

	mock.ExpectQuery("SELECT id, note, duedate, repeat, completed FROM todo").
		WillReturnRows(rows)

	return db, mock, todo.List{}
}

// GetItemSetupHelper encapsulates common code needed to setup mock DB access a single todo item's data
func GetItemSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *todo.Item) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "note", "duedate", "repeat", "completed"}).
		AddRow(1, "Get groceries", now, false, false)

	mock.ExpectQuery("SELECT id, note, duedate, repeat, completed FROM todo").
		WillReturnRows(rows)

	expected := todo.Item{
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
func DBCallNoExpectationsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *todo.Item) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
	}

	return db, mock, nil
}

// // DBNoCallSetupHelper encapsulates the common code needed to mock an error upstream from an actual DB call
// func DBNoCallSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	return db, mock
// }

// 	rows := sqlmock.NewRows([]string{"accountid", "id", "name", "email", "role"}).
// 		AddRow(5, 1, "porgy tirebiter", "porgytirebiter@email.com", todo.Primary)

// 	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM todo").
// 		WithArgs(1).WillReturnRows(rows)

// 	expected := todo.Item{
// 		AccountID: 5,
// 		ID:        1,
// 		Name:      "porgy tirebiter",
// 		EMail:     "porgytirebiter@email.com",
// 		Role:      todo.Primary,
// 	}

// 	return db, mock, &expected
// }

// // DBUserErrNoRowsSetupHelper encapsulates common coded needed to mock Queries returning no rows
// func DBUserErrNoRowsSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *todo.Item) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM todo").
// 		WithArgs(1).WillReturnError(sql.ErrNoRows)

// 	return db, mock, nil
// }

// // DBUserOtherErrSetupHelper encapsulates common coded needed to mock Queries returning no rows
// func DBUserOtherErrSetupHelper(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *todo.Item) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM todo").
// 		WithArgs(1).WillReturnError(sql.ErrConnDone)

// 	return db, mock, nil
// }

//
//
//
//
//
//
//
//

func validateExpectedErrors(t *testing.T, err error, shouldPass bool) {
	if shouldPass && err != nil {
		t.Fatalf("error '%s' was not expected", err)
	}
	if !shouldPass && err == nil {
		t.Fatalf("expected error didn't occur")
	}
}

// // DBDeleteSetupHelper encapsulates the common code needed to setup a mock Item delete
// func DBDeleteSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectExec("DELETE FROM todo WHERE id = ?").WithArgs(u.ID).
// 		WillReturnResult(sqlmock.NewResult(0, 1))

// 	return db, mock
// }

// // DBDeleteErrorSetupHelper encapsulates the common code needed to mock a todo delete error
// func DBDeleteErrorSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectExec("DELETE FROM todo WHERE id = ?").WithArgs(u.ID).WillReturnError(sql.ErrConnDone)

// 	return db, mock
// }

// // DBInsertSetupHelper encapsulates the common code needed to setup a mock Item insert
// func DBInsertSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectExec("INSERT INTO todo").WithArgs(u.AccountID, u.Name, u.EMail, u.Role, u.Password).
// 		WillReturnResult(sqlmock.NewResult(1, 1))

// 	return db, mock
// }

// // DBInsertErrorSetupHelper encapsulates the common code needed to mock a todo insert error
// func DBInsertErrorSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectExec("INSERT INTO todo").WithArgs(u.AccountID, u.Name, u.EMail, u.Role, u.Password).
// 		WillReturnError(fmt.Errorf("some error"))

// 	return db, mock
// }

// // DBUpdateNonExistingRowSetupHelper mimics an update to a non-existing todo, can't update non-existing todo.
// func DBUpdateNonExistingRowSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	rows := sqlmock.NewRows([]string{"accountid", "id", "name", "email", "role"}).
// 		AddRow(1, 100, "Mickey Mouse", "MickeyMoused@disney.com", todo.Unrestricted)

// 	mock.ExpectBegin()
// 	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM todo WHERE id = ?").WithArgs(u.ID).
// 		WillReturnRows(rows)

// 	return db, mock
// }

// // DBUpdateErrorSelectSetupHelper mimics an update where the non-existence query fails.
// func DBUpdateErrorSelectSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectBegin()
// 	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM todo WHERE id = ?").WithArgs(u.ID).
// 		WillReturnError(sql.ErrConnDone)

// 	return db, mock
// }

// // DBUpdateSetupHelper encapsulates the common code needed to setup a mock Item update
// func DBUpdateSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectBegin()
// 	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM todo WHERE id = ?").WithArgs(u.ID).WillReturnError(sql.ErrNoRows)
// 	mock.ExpectExec("UPDATE todo SET (.+) WHERE (.+)").WithArgs(u.ID, u.AccountID, u.Name, u.EMail, u.Role, u.Password, u.ID).
// 		WillReturnResult(sqlmock.NewResult(0, 1)) // no insert ID, 1 row affected
// 	// mock.ExpectExec("UPDATE todo").WithArgs(sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg, sqlmock.AnyArg).
// 	// 	WillReturnResult(sqlmock.NewResult(0, 1)) // no insert ID, 1 row affected
// 	mock.ExpectCommit()
// 	return db, mock
// }

// // DBUpdateErrorSetupHelper encapsulates the common code needed to setup a mock Item update error
// func DBUpdateErrorSetupHelper(t *testing.T, u todo.Item) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a mock database connection", err)
// 	}

// 	mock.ExpectBegin()
// 	mock.ExpectQuery("SELECT accountID, id, name, email, role FROM todo WHERE id = ?").WithArgs(u.ID).WillReturnError(sql.ErrNoRows)
// 	mock.ExpectExec("UPDATE todo SET (.+) WHERE (.+)").WithArgs(u.ID, u.AccountID, u.Name, u.EMail, u.Role, u.Password, u.ID).WillReturnError(sql.ErrConnDone)
// 	mock.ExpectRollback()
// 	return db, mock
// }
