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
	getToDoQuery     = "SELECT id, note, duedate, repeat, completed FROM todo WHERE id = $1"
	insertToDoStmt   = "INSERT INTO todo (note, duedate, repeat, completed) VALUES ($1, $2, $3, $4) RETURNING id"
	updateToDoStmt   = "UPDATE todo SET note = $1, duedate = $2, repeat = $3, completed = $4 WHERE id = $5"
	deleteToDoStmt   = "DELETE FROM todo WHERE id = $1"
)

// Item represents the data about a To Do list item
type Item struct {
	ID        int64     `json:"id"`
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
func GetToDoList(db *sql.DB) (List, error) {
	results, err := db.Query(getAllToDosQuery)
	if err != nil {
		return List{}, errors.Annotate(err, "error querying DB")
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
			return List{}, errors.Annotate(err, "error scanning result set")
		}

		tdl.Items = append(tdl.Items, &td)
	}

	return tdl, nil
}

// GetToDoItem will return the todo identified by 'id' or a nil todo if there
// wasn't a matching todo.
func GetToDoItem(db *sql.DB, id int) (*Item, error) {
	row := db.QueryRow(getToDoQuery, id)
	var td Item
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

	return &td, nil
}

// InsertToDo takes the provided todo data, inserts it into the db, and returns the newly created todo ID.
func InsertToDo(db *sql.DB, td Item) (int64, error) {
	err := validateToDo(td)
	if err != nil {
		return 0, errors.Annotate(err, "ToDo validation failure")
	}

	var id int64
	err = db.QueryRow(insertToDoStmt, td.Note, td.DueDate, td.Repeat, td.Completed).Scan(&id)
	if err != nil {
		return 0, errors.Annotate(err, fmt.Sprintf("error inserting todo %+v into DB", td))
	}

	return id, nil
}

// UpdateToDo takes the provided todo data, inserts it into the db, and returns the newly created todo ID
func UpdateToDo(db *sql.DB, td Item) (constants.ErrCode, error) {
	err := validateToDo(td)
	if err != nil {
		return constants.DBUpSertErrorCode, errors.Annotate(err, "ToDo validation failure")
	}

	_, err = db.Exec(updateToDoStmt, td.Note, td.DueDate, td.Repeat, td.Completed, td.ID)
	if err != nil {
		return constants.DBUpSertErrorCode, errors.Annotate(err, fmt.Sprintf("error updating todo in the database: %+v", td))
	}

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
