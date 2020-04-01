DROP TABLE IF EXISTS todo;
CREATE TABLE todo (
    id SERIAL PRIMARY KEY,
    note text,
    dueDate timestamp,
    repeat boolean DEFAULT false,
    completed boolean DEFAULT false
);
