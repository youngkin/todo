DROP DATABASE IF EXISTS todo;
DROP USER IF EXISTS todo;

CREATE USER todo WITH PASSWORD 'todo123';
CREATE DATABASE todo WITH OWNER todo;