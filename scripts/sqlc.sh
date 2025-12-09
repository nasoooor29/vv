#!/bin/bash

TYPES_FILE="./src/types/types.gen.ts"
ZOD_TYPES_FILE="./src/types/zod-types.gen.ts"

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

generate_types() {
    tygo gendir ./internal -r -o "$TYPES_FILE" >/dev/null
    cd ./frontend || exit
    bun i prettier ts-to-zod -D
    print_blue "converting types to zod schemas..."
    bunx ts-to-zod "$TYPES_FILE" -o "$ZOD_TYPES_FILE"
    print_blue "zod schemas generated"
}

format_and_cleanup() {
    gofumpt -w ./internal
    gofumpt -w ./cmd
    bunx prettier --write "$TYPES_FILE" "$ZOD_TYPES_FILE"
    print_blue "formatted types and zod schemas"
    cd ../.. || exit
}

main() {
    generate_sqlc || exit 1
    generate_types
    format_and_cleanup
}

main
