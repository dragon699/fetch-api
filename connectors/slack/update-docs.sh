#!/bin/bash


export DIR="./internal,./cmd/slack"
export GENERAL_INFO="./app/bootstrap.go"
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
