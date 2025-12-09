#!/bin/bash

set -e

TYPES_FILE="./frontend/src/types/types.gen.ts"
ZOD_TYPES_FILE="./frontend/src/types/zod-types.gen.ts"

print_blue() {
    echo -e "\033[34m$1\033[0m"
}

print_red() {
    echo -e "\033[31m$1\033[0m"
}

generate_sqlc() {
    sqlc generate
    if [ $? -ne 0 ]; then
        print_red " sqlc generate failed "
        return 1
    fi
    print_blue "types generated"
}
fix_some_types() {
    local types_file=$1
    sed -i 's/any \/[*] time\.Time [*]\//Date/g' $types_file
}

generate_types() {
    # variable called files will contain all the go files in the internal/models and internal/database and remove any file called models.go
    local FILES=$(find ./internal/models ./internal/database -name "*.go" ! -name "models.go" ! -name "db.go")
    tygo gendir $FILES ./internal/database/user/models.go -r -o "$TYPES_FILE" >/dev/null
    print_blue "fixing some types..."
    fix_some_types "$TYPES_FILE"
    cd ./frontend || exit
    bun i prettier ts-to-zod -D
    print_blue "converting types to zod schemas..."
    # paths are relative to the frontend directory
    bunx ts-to-zod ./src/types/types.gen.ts ./src/types/zod-types.gen.ts --skipValidation
    print_blue "zod schemas generated"
    cd .. || exit
}

format_and_cleanup() {
    gofumpt -w ./internal
    gofumpt -w ./cmd
    bunx prettier --write "$TYPES_FILE" "$ZOD_TYPES_FILE"
    print_blue "formatted types and zod schemas"
}

main() {
    generate_sqlc || exit 1
    generate_types
    format_and_cleanup
}

main
