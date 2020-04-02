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
    duedate: {string}    // Time/date 
    repeat: {bool}       // Valid values are 'true' or 'false'
    completed: {bool}    // Valid values are 'true' or 'false'
}
```

Example:

``` JSON
{
    id: 1,
    note: "Get groceries",
    duedate: "2020-04-01T00:00:00Z",
    repeat: false,
    completed: false,
}
```

## Resources

|Verb   | Resource | Description  | Status  | Status Description |
|:------|:---------|:-------------|--------:|:-------------------|
|GET    |/health   |Health check, returns `I'm Healthy!` if all's OK| 200| Service healthy |
|GET    |/todo     |Get all To Do items, do not include `id` in JSON body| 200|All To Do items returned |
|GET    |/todo/{id}|Get the To Do item identified by {id}| 200|To Do item returned |
|       |          |              | 404| To do item not found|
|POST   |/todo     |Create a new To Do item, do not include `id` in JSON body| 201|To Do item successfully created|
|PUT    |/todo/{id}|Update an existing To Do item identified by {id}, pass complete JSON in body|200|To Do item updated|
|       |          |              | 404| To do item not found|
|DELETE |/todo/{id}|Deletes the referenced resource|200|To Do item was deleted|
|       |          |                               |404|To Do item was not found|

## Common HTTP status codes

|Status|Action|
|-----:|:-----|
|400|Bad request, don't retry|
|429|Server busy, can retry after `Retry-After` time has expired (in seconds)|
|500|Internal server error, can retry|

# Runnning and testing the application

## Pre-commit check

From the project root directory (`todoshaleapps`) run:

``` bash
./precheck
```

This runs `go vet ./...`, `go fmt ./...`, `golint ./...`, and `go test -race ./...`

## Running the application

The host address in the example URLs references a deployment in Google Kubernetes Engine in my personal account. This can be used when the application isn't deployed elsewhere (e.g., locally or on another Kubernetes cluster).

It's possible to deploy the application to another Kubernetes cluster using the Kubernetes specs in `todoshaleapps/kubernetes`.

The application can also be run locally by:

1. Building an executable by running `go  build` in `todoshaleapps/src/cmd/todod`.
2. Start the `todod` service using this command line (again from `todoshaleapps/src/cmd/todod`):

``` bash
./todod -dbport <postgres port number> -dbhost <postgres IP address> -dbuser <postgres user ID> -dbpassword <postgres user password>
```
   If the database name isn't configured to be `todo` as described below, an additional command line flag, `-dbname`, can be provided.

In these alternate deployments the host IP address in the examples should be modified to reflect the correct location. A Postgres database will also need to be available. The following changes will have to made to reference the Postgres database:

1. From `todoshaleapps/sql`
   1. Log into Postgres (`psql`)
   2. Create a database called `todo`
   3. Change to the `todo` database (`\c todo`)
   4. Run `\i createtables.sql`
   5. Run `\i testdata.sql`

   This will create the required tables as well as populate them with some sample data

The application logs to `stdout`:

```
kt logs todod-84f9c7788-q9g9h
{"Application":"ToDo","HostName":"todod-84f9c7788-q9g9h","LogLevel":"info","Port":":8080","level":"info","msg":"todod service starting","time":"2020-04-02T19:18:59Z"}

{"Application":"ToDo","HostName":"todod-84f9c7788-q9g9h","ServiceName":"health","level":"info","msg":"handling request","time":"2020-04-02T19:19:00Z"}

{"Application":"ToDo","HTTPMethod":"GET","HostName":"todod-84f9c7788-q9g9h","RemoteAddr":"10.8.0.1:53917","URLPath":"/todos","level":"info","msg":"HTTP request received","time":"2020-04-02T19:19:54Z"}

{"Application":"ToDo","HTTPMethod":"POST","HostName":"todod-84f9c7788-q9g9h","RemoteAddr":"10.8.0.1:54051","URLPath":"/todos","level":"info","msg":"HTTP request received","time":"2020-04-02T19:20:44Z"}

{"Application":"ToDo","HTTPMethod":"PUT","HostName":"todod-84f9c7788-q9g9h","RemoteAddr":"10.8.0.1:54219","URLPath":"/todos/6","level":"info","msg":"HTTP request received","time":"2020-04-02T19:23:10Z"}

{"Application":"ToDo","ErrorDetail":"resource ID in url (7) doesn't match resource ID in request body (6)","HTTPStatus":400,"HostName":"Richs-MacBook.local","URLPath":"/todos/7","level":"error","msg":"1000","time":"2020-04-02T14:43:40-06:00"}

{"Application":"ToDo","HTTPMethod":"DELETE","HostName":"todod-7f47847987-fjlk2","RemoteAddr":"10.8.0.1:64925","URLPath":"/todos/7","level":"info","msg":"HTTP request received","time":"2020-04-02T20:48:50Z"}
```

## Example `curl` commands

* Get a To Do List

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

* Get a To Do Item
  
```
curl http://35.227.143.9:80/todos/3 | jq "."
{
  "id": 3,
  "selfref": "/todos/3",
  "note": "walk dog",
  "duedate": "2020-04-03T12:00:00Z",
  "repeat": true,
  "completed": false
}
```

* Create a new To Do Item

```
curl -i -X POST http://35.227.143.9:80/todos -H "Content-Type: application/json" -d "{\"note\":\"work out\",\"duedate\":\"2020-04-01T00:00:00Z\",\"repeat\":true,\"completed\":false}"
HTTP/1.1 201 Created
Location: /todos/6
Date: Thu, 02 Apr 2020 02:49:08 GMT
Content-Length: 0
```

* Update To Do Item

```
curl -i -X PUT http://35.227.143.9:80/todos/4 -H "Content-Type: application/json" -d "{\"id\":4,\"note\":\"workout extra hard\",\"duedate\":\"2525-04-02T13:13:13Z\",\"repeat\":true,\"completed\":true}"
HTTP/1.1 200 OK
Date: Thu, 02 Apr 2020 18:29:45 GMT
Content-Length: 0

```

* Delete a To Do Item

```
curl -i -X DELETE http://35.227.143.9:80/todos/7
HTTP/1.1 200 OK
Date: Thu, 02 Apr 2020 20:48:50 GMT
Content-Length: 0
```

# Things I would have liked to have had working

I intended to write unit tests against a mocked SQL database using `go-sqlmock`. I have succesfully used this mocking framework in the past for just this type of thing. In those instances though I was using the MySQL DB driver. There is a slight difference in the structure of SQL statements between MySQL and Postgres as well as differences between what MySQL and Postgres return from `INSERT`s, `UPDATE`s and `DELETE`s. At this point I'm wondering if there's an issue with trying to mock Postgres. I intend to look into this a bit more, but for now there are only unit tests for `SELECT` DB requests.

# Future Enhancements

1. Support `context` in DB calls
2. Support partial updates in `PUT` requests
3. Support for multiple users
4. Support for metrics (e.g., Prometheus)
5. Create automated integration tests (i.e., tests against the full application)
