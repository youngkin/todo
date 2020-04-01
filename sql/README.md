# SQL Overview

There are 2 files containing SQL DDL/DML commands:

1. `createtables.sql` - As the name indicates, this file contains the commands to create the table(s) comprising the To Do application
2. `testdata.sql` - This file contains the `INSERT` commands to create a simple set of test data.

To use these files:

1. Log into the postgres server: 

```
psql -h <somehost> -p <someport> -U <someuser> --password <dbname>
```

From the `psql` command line run:

```
\i createtables.sql
\i testdata.sql
```

# Database description

The database for the app is called `todo`. The table used to represent a todo item is named `todo` as well. The columns are as follows:

```
'note' is the text of the To Do item (e.g., get groceries)
'dueDate' is the date/time when the To Do item should be complete
'repeat' indicates if the item will be repeated daily until due date
'completed' indicates if the item has been completed , 'true' if it has, 'false' if not.
```
