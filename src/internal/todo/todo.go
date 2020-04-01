package todo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/youngkin/todoshaleapps/src/internal/platform/constants"
)

var (
	getAllToDosQuery = "SELECT id, note, duedate, repeat, completed FROM todo"
	getToDoQuery     = "SELECT id, note, duedate, repeat, completed FROM todo WHERE id = ?"
	insertToDoStmt   = "INSERT INTO todo (note, duedate, repeat, completed) VALUES (?, ?, ?, ?)"
	updateToDoStmt   = "UPDATE todo SET note = ?, duedate = ?, repeat = ?, completed = ? WHERE id = ?"
	deleteToDoStmt   = "DELETE FROM todo WHERE id = ?"
)

// Item represents the data about a To Do list item
type Item struct {
	ID        int       `json:"id"`
	SelfRef   string    `json:"selfref"`
	Note      string    `json:"note"`
	DueDate   time.Time `json:"duedate"`
	Repeat    bool      `json:"repeat"`
	Completed bool      `json:"completed"`
}

// List is a collection ToDo items, i.e., a To Do List
type List struct {
	Items []*Item `json:"todolist"`
}

// GetToDoList will return all ToDo items
func GetToDoList(db *sql.DB) (*List, error) {
	results, err := db.Query(getAllToDosQuery)
	if err != nil {
		return &List{}, errors.Annotate(err, "error querying DB")
	}

	tdl := List{}
	for results.Next() {
		var td Item

		err = results.Scan(&td.ID,
			&td.Note,
			&td.DueDate,
			&td.Repeat,
			&td.Completed)
		if err != nil {
			return &List{}, errors.Annotate(err, "error scanning result set")
		}

		tdl.Items = append(tdl.Items, &td)
	}

	return &tdl, nil
}

// GetToDoItem will return the todo identified by 'id' or a nil todo if there
// wasn't a matching todo.
func GetToDoItem(db *sql.DB, id int) (*Item, error) {
	row := db.QueryRow(getToDoQuery, id)
	td := &Item{}
	err := row.Scan(&td.ID,
		&td.Note,
		&td.DueDate,
		&td.Repeat,
		&td.Completed)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Annotate(err, "error scanning todo row")
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return td, nil
}

// InsertToDo takes the provided todo data, inserts it into the db, and returns the newly created todo ID.
func InsertToDo(db *sql.DB, td Item) (int64, constants.ErrCode, error) {
	err := validateToDo(td)
	if err != nil {
		return 0, constants.ToDoValidationErrorCode, errors.Annotate(err, "ToDo validation failure")
	}

	r, err := db.Exec(insertToDoStmt, td.Note, td.DueDate, td.Repeat, td.Completed)
	if err != nil {
		return 0, constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error inserting todo %+v into DB", td))
	}
	id, err := r.LastInsertId()
	if err != nil {
		return 0, constants.DBUpSertErrorCode, errors.Annotate(err, "error getting todo ID")
	}

	return id, constants.NoErrorCode, nil
}

// UpdateToDo takes the provided todo data, inserts it into the db, and returns the newly created todo ID
func UpdateToDo(db *sql.DB, td Item) (constants.ErrCode, error) {
	err := validateToDo(td)
	if err != nil {
		return constants.DBUpSertErrorCode, errors.Annotate(err, "ToDo validation failure")
	}

	// TODO: Confirm this statement
	// This entire db.Begin/tx.Rollback/Commit seem awkward to me. But it's here because
	// Postgres silently performs an insert if there is no row to update.
	tx, err := db.Begin()
	if err != nil {
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error beginning transaction for todo: %+v", td))
	}
	row := db.QueryRow(getToDoQuery, td.ID)
	err = row.Scan(&td.ID,
		&td.Note,
		&td.DueDate,
		&td.Repeat,
		&td.Completed)
	if err == nil {
		return constants.DBInvalidRequestCode, errors.New(fmt.Sprintf("error, attempting to update non-existent todo, todo.ID %d", td.ID))
	}
	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error updating todo in the database: %+v", td))
	}

	_, err = db.Exec(updateToDoStmt, td.Note, td.DueDate, td.Repeat, td.Completed, td.ID)
	if err != nil {
		tx.Rollback()
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error updating todo in the database: %+v", td))
	}
	tx.Commit()

	return constants.NoErrorCode, nil
}

// DeleteToDo deletes the todo identified by td.id from the database
func DeleteToDo(db *sql.DB, id int) (constants.ErrCode, error) {
	_, err := db.Exec(deleteToDoStmt, id)
	if err != nil {
		return constants.DBDeleteErrorCode, errors.Annotate(err, fmt.Sprintf("ToDo delete error for ID %d", id))
	}

	return constants.NoErrorCode, nil
}

func validateToDo(td Item) error {
	errMsg := ""

	if len(td.Note) == 0 {
		errMsg = errMsg + "ToDo note must be populated"
	}

	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}
	return nil
}
