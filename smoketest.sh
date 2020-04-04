#!/bin/bash
# Smoke tests for ToDo app
#
if [  $1 = "help" ]
then
    echo "usage:"
    echo "    smoketest <dbaddr> <dbport> <svcaddr> <svcport>"
    echo "Example:"
    echo "    smoketest localhost 5432 localhost 8080"
    exit
fi

DBAddr=$1
DBPort=$2
ToDoAddr=$3
ToDoPort=$4
export PGPASSWORD="todo123"


echo "Create table"
psql -h $DBAddr -p $DBPort -U todo todo -f sql/createtables.sql -a
echo ""
echo "Pre-populate table with test data"
psql -h $DBAddr -p $DBPort -U todo todo -f sql/testdata.sql -a

echo ""
echo "curl http://"$ToDoAddr":"$ToDoPort"todos"
echo "Get To Do List"
curl http://$ToDoAddr:$ToDoPort/todos | jq "."
echo ""
echo "Get To Do item 3"
curl http://$ToDoAddr:$ToDoPort/todos/3 | jq "."
echo ""
echo "Add a To Do item for working out. Should get a 201 and the 'Location' header should have '/todos/4'"
curl -i -X POST http://$ToDoAddr:$ToDoPort/todos -H "Content-Type: application/json" -d "{\"note\":\"work out\",\"duedate\":\"2020-04-01T00:00:00Z\",\"repeat\":true,\"completed\":false}"
echo ""
echo "Get new To Do item"
curl http://$ToDoAddr:$ToDoPort/todos/4 | jq "."
echo ""
echo "Update To Do to 'workout extra hard'"
curl -i -X PUT http://$ToDoAddr:$ToDoPort/todos/4 -H "Content-Type: application/json" -d "{\"id\":4,\"note\":\"workout extra hard\",\"duedate\":\"2525-04-02T13:13:13Z\",\"repeat\":true,\"completed\":true}"
echo ""
echo "Look at the update"
curl http://$ToDoAddr:$ToDoPort/todos/4 | jq "."
echo ""
echo "Delete the newly added To Do item. Should get a 200."
curl -i -X DELETE http://$ToDoAddr:$ToDoPort/todos/4
echo ""
echo "Is the deleted item still there? Should get a 404"
curl -i http://$ToDoAddr:$ToDoPort/todos/4
echo ""
echo "Back to the original To Do List"
curl http://$ToDoAddr:$ToDoPort/todos | jq "."
echo ""
echo "Do a bulk POST with results of operation in body of response"
curl -X POST http://$ToDoAddr:$ToDoPort/todos?bulk=true -H "Content-Type: application/json" -d "{\"todolist\": [{\"note\": \"get groceries\",\"duedate\": \"2020-04-01T00:00:00Z\",\"repeat\": false,\"completed\": false},{\"note\": \"pay bills\",\"duedate\": \"2020-04-02T00:00:00Z\",\"repeat\": false,\"completed\": false},{\"note\": \"walk dog\",\"duedate\": \"2020-04-03T12:00:00Z\",\"repeat\": true,\"completed\": false}]}" | jq "."
echo ""
echo "Should now have 6 To Do items in the list"
curl http://$ToDoAddr:$ToDoPort/todos | jq "."


echo ""
echo ""
echo ""
echo "Running failure tests..."
echo ""
echo ""
echo ""
echo "POST an item with an invalid date, should return a 400"
curl -i -X POST http://$ToDoAddr:$ToDoPort/todos -H "Content-Type: application/json" -d "{\"note\":\"work out\",\"duedate\":\"JULY 5TH\",\"repeat\":true,\"completed\":false}"
echo ""
echo "POST an item with an invalid bool value, should return a 400"
curl -i -X POST http://$ToDoAddr:$ToDoPort/todos -H "Content-Type: application/json" -d "{\"note\":\"work out\",\"duedate\":\"2020-04-02T13:13:13Z\",\"repeat\":true,\"completed\":NUITSNUT}"
echo ""
echo "POST an item with an invalid JSON tag, should return a 400"
curl -i -X POST http://$ToDoAddr:$ToDoPort/todos -H "Content-Type: application/json" -d "{\"BADNOTE\":\"work out\",\"duedate\":\"2020-04-02T13:13:13Z\",\"repeat\":true,\"completed\":false}"
echo ""
echo "DELETE an item with no {ID}, should return a 400"
curl -i -X DELETE http://$ToDoAddr:$ToDoPort/todos/
echo ""
echo "DELETE an item with an invalid {ID}, should return a 400"
curl -i -X DELETE http://$ToDoAddr:$ToDoPort/todos/BADID
echo ""
echo "DELETE a non-existing item. Will return a 200 OK mirroring Postgres behavior"
curl -i -X DELETE http://$ToDoAddr:$ToDoPort/todos/2001
echo ""
echo "Send a request with an unsupported HTTP Verb, should return a 501"
curl -i -X PATCH http://$ToDoAddr:$ToDoPort/todos/2
echo ""
echo "POST a bulk request with an invalided sub-request. Should return a 409 with the problem clearly identified in the response."
curl -i -X POST http://$ToDoAddr:$ToDoPort/todos?bulk=true -H "Content-Type: application/json" -d "{\"todolist\": [{\"id\":1,\"note\": \"get groceries\",\"duedate\": \"2020-04-01T00:00:00Z\",\"repeat\": false,\"completed\": false},{\"note\": \"pay bills\",\"duedate\": \"2020-04-02T00:00:00Z\",\"repeat\": false,\"completed\": false},{\"note\": \"walk dog\",\"duedate\": \"2020-04-03T12:00:00Z\",\"repeat\": true,\"completed\": false}]}"
