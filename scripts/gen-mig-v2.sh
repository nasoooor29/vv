#!/bin/bash

# https://github.com/nasoooor29/mi-py

# if the user did not provide a name for the migration, ask him to provide one
if [ -z "$1" ]; then
    echo "Please provide a name for the migration"
    # take input from user
    read -p "Migration name: " MIGRATION_NAME
else
    MIGRATION_NAME=$1
fi
# check if sql-differ exists
if ! command -v sql-differ &> /dev/null
then
    echo "sql-differ could not be found, please install it first."
    exit 1
fi
TS=$(date -u +"%Y%m%d%H%M%S") 

sql-differ migrate ./visory.db ./internal/database/visory.sql "./internal/database/migrations/${TS}_${MIGRATION_NAME}"

# # Generate visory.sql from the database
# goose -dir ./internal/database/migrations sqlite3 visory.db up
# sql-differ generate ./visory.db ./internal/database/visory.sql
