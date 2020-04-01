[![Build Status](https://travis-ci.org/youngkin/todoshaleapps.svg?branch=master)](https://travis-ci.org/youngkin/todoshaleapps) [![Go Report Card](https://goreportcard.com/badge/github.com/youngkin/todoshaleapps)](https://goreportcard.com/report/github.com/youngkin/todoshaleapps)

# todoshaleapps

Service for a simple ToDo application

# Database

The database for the app is called `todo`. The table used to represent a todo item is named `todo` as well. The columns are as follows:

```
'note' is the text of the To Do item (e.g., get groceries)
'dueDate' is the date/time when the To Do item should be complete
'repeat' indicates if the item will be repeated daily until due date
'completed' indicates if the item has been completed , 'true' if it has, 'false' if not.
```
