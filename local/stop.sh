#!/bin/bash


DOCKER_COMPOSE_FILE="./docker-compose.yaml"


function stop() {
    STOP_CMD="sudo docker compose -f ${DOCKER_COMPOSE_FILE} down"

    echo " > Stopping containers.."
    ${STOP_CMD}
}

stop
