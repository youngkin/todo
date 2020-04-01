[![Build Status](https://travis-ci.org/youngkin/todoshaleapps.svg?branch=master)](https://travis-ci.org/youngkin/todoshaleapps) [![Go Report Card](https://goreportcard.com/badge/github.com/youngkin/todoshaleapps)](https://goreportcard.com/report/github.com/youngkin/todoshaleapps)

# todoshaleapps

Service for a simple application that manages a To Do list

# API

## Representation

A To Do item is represented in JSON as follows:

``` 
{
    id: {int}            // Resource identifier, don't populate on POST
    selfref: {string}    // Resource URL, e.g., /todo/1. Returned on GET. Don't populate for POST/PUT
    note: {string}
    duedate: {string}    // Time/date + timezone offset (e.g., +0 for GMT)
    repeat: {bool}       // Valid values are 'true' or 'false'
    completed: {bool}    // Valid values are 'true' or 'false'
}
```

Example:

``` JSON
{
    id: 1,
    note: "Get groceries",
    duedate: "04-01-2020 12:00:00+0",
    repeat: false,
    completed: false,
}
```

## Resources

|Verb   | Resource | Description  | Status  | Status Description |
|:------|:---------|:-------------|--------:|:-------------------|
|GET    |/health   |Health check, returns "I'm Healthy!" if all's OK| 200| Service healthy |
|GET    |/todo     |Get all To Do items, do not include `id` in JSON body| 200|All To Do items returned |
|GET    |/todo/{id}|Get the To Do item identified by {id}| 200|To Do item returned |
|       |          |              | 404| To do item not found|
|POST   |/todo     |Create a new To Do item| 201|To Do item successfully created|
|PUT    |/todo/{id}|Update an existing To Do item identified by {id}, pass complete JSON in body|200|To Do item updated|
|       |          |              | 404| To do item not found|

## Common HTTP status codes

|Status|Action|
|-----:|:-----|
|400|Bad request, don't retry|
|429|Server busy, can retry after `Retry-After` time has expired (in seconds)|
|500|Internal server error, can retry|

## Examples

```
curl http://35.227.143.9:80/todos | jq "."
{
  "todolist": [
    {
      "id": 1,
      "selfref": "/todos/1",
      "note": "get groceries",
      "duedate": "2020-04-01T00:00:00Z",
      "repeat": false,
      "completed": false
    },
    {
      "id": 2,
      "selfref": "/todos/2",
      "note": "pay bills",
      "duedate": "2020-04-02T00:00:00Z",
      "repeat": false,
      "completed": false
    },
    {
      "id": 3,
      "selfref": "/todos/3",
      "note": "walk dog",
      "duedate": "2020-04-03T12:00:00Z",
      "repeat": true,
      "completed": false
    }
  ]
}
```

# Future Enhancements

1. Support `context` in DB calls
2. Support partial updates in `PUT` requests
3. Support for multiple users
