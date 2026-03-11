#!/bin/bash


DOCKER_COMPOSE_FILE="./docker-compose.yaml"
DETACHED=false


function run() {
    BUILD_CMD="sudo docker compose -f ${DOCKER_COMPOSE_FILE} build --no-cache"
    RUN_CMD="sudo docker compose -f ${DOCKER_COMPOSE_FILE} up"

    if [[ "${DETACHED}" == true ]]; then
        RUN_CMD="${RUN_CMD} -d"
    fi

    echo " > Re-building with latest source code.."
    ${BUILD_CMD}

    echo " > Starting containers.."
    ${RUN_CMD}
}

run
