#!/bin/bash


# swag only resolves types from the listed package dirs; "./" has no Go files
export DIR="./internal/app,./internal/http/routes,./internal/http/dto/response,./internal/http/dto/request"
export GENERAL_INFO="bootstrap.go"
export OUTPUT_DIR="./docs"


function update_docs() {
    $(go env GOPATH)/bin/swag init \
    --parseDependency \
    --parseInternal \
    --dir "${DIR}" \
    --generalInfo "${GENERAL_INFO}" \
    --output "${OUTPUT_DIR}"
}

echo "Updating Swagger docs..."
update_docs
echo "Updated Swagger docs"
