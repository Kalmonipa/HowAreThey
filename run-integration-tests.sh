#!/bin/bash
set -o errexit

BRANCH_NAME=$(git branch --show-current)
IMAGE_NAME="kalmonipa/howarethey:$BRANCH_NAME"
CONTAINER_NAME="basic-test"

## Clean up
cleanup() {
    echo "Cleaning up $CONTAINER_NAME"
    docker stop "$CONTAINER_NAME"
}

trap cleanup EXIT

## Build the image to test
#docker build . --tag "$IMAGE_NAME"

## Run a basic test
docker run --rm -p 8080:8080 --detach --name "$CONTAINER_NAME" "$IMAGE_NAME"

sleep 5

ACTUAL_RESPONSE=$(curl --silent --request POST http://localhost:8080/friends \
     --data "{\"Name\":\"Jack Reacher\",\"Birthday\":\"29/10/1960\"}")

EXPECTED_RESPONSE='{"message":"Jack Reacher added successfully"}'

if [[ $ACTUAL_RESPONSE != $EXPECTED_RESPONSE ]]; then
    echo "ERROR: POST request to add a friend failed"
    echo "RESPONSE: $ACTUAL_RESPONSE"
    exit 1
else
    echo "INFO: POST request to add a friend passed"
fi

ACTUAL_RESPONSE=$(curl --silent http://localhost:8080/friends)

EXPECTED_RESPONSE='[{"ID":"1","Name":"Jack Reacher","LastContacted":"","Birthday":"29/10/1960","Notes":""}]'

# TODO: Find out why doing a direct comparison fails the tests...
if diff <(echo "$EXPECTED_RESPONSE" | tr -d '[:space:]') <(echo "$ACTUAL_RESPONSE" | tr -d '[:space:]') >/dev/null; then
    echo "INFO: GET request to retrieve friends passed"
else
    echo "ERROR: GET request to retrieve friends failed"
    echo "  expected: $EXPECTED_RESPONSE"
    echo "       got: $ACTUAL_RESPONSE"
    exit 1
fi


#sleep 30
