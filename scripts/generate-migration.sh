#!/bin/bash

# if the user did not provide a name for the migration, ask him to provide one
if [ -z "$1" ]; then
    echo "Please provide a name for the migration"
    # take input from user
    read -p "Migration name: " MIGRATION_NAME
else
    MIGRATION_NAME=$1
fi
echo "Creating migration: $MIGRATION_NAME"
goose create -dir ./internal/database/migrations "$MIGRATION_NAME" sql

# Generate visory.sql from the database
goose -dir ./internal/database/migrations sqlite3 visory.db up
sqlite3 visory.db .schema > ./internal/database/visory.sql
