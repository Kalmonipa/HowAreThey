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

build_and_run() {
    ## Build the image to test
    #docker build . --tag "$IMAGE_NAME"

    docker run --rm -p 8080:8080 --detach --name "$CONTAINER_NAME" "$IMAGE_NAME"

    sleep 5
}

compare_responses() {
    ACTUAL_RESPONSE=$1
    EXPECTED_RESPONSE=$2
    TEST_DESCRIPTION=$3

    if [[ $ACTUAL_RESPONSE != $EXPECTED_RESPONSE ]]; then
        echo "ERROR: $TEST_DESCRIPTION failed"
        echo "  expected: $EXPECTED_RESPONSE"
        echo "       got: $ACTUAL_RESPONSE"
        exit 1
    else
        echo "INFO: $TEST_DESCRIPTION passed"
    fi
}

## Function to add a friend
add_friends() {
    TODAYS_DATE=$(date '+%d/%m/%Y')

    ACTUAL_RESPONSE=$(curl --silent --request POST \
        --write-out "%{http_code}\n" \
        --output /dev/null \
        --data "{\"Name\":\"Jack Reacher\",\"LastContacted\":\"06/06/2023\",\"Birthday\":\"$TODAYS_DATE\"}" \
        http://localhost:8080/friends)

    EXPECTED_RESPONSE='201'

    # Checking to make sure the friend addition worked properly
    compare_responses $ACTUAL_RESPONSE $EXPECTED_RESPONSE "POST request to add a friend"
}

## Test functions



main() {

    add_friends

    ## Test GET /friends
    ACTUAL_RESPONSE=$(curl --silent http://localhost:8080/friends)

    EXPECTED_RESPONSE="[{\"ID\":\"1\",\"Name\":\"Jack Reacher\",\"LastContacted\":\"06/06/2023\",\"Birthday\":\"$TODAYS_DATE\",\"Notes\":\"\"}]"

    # TODO: Find out why doing a direct comparison fails the tests...
    if diff <(echo "$EXPECTED_RESPONSE" | tr -d '[:space:]') <(echo "$ACTUAL_RESPONSE" | tr -d '[:space:]') >/dev/null; then
        echo "INFO: GET request to retrieve friends passed"
    else
        echo "ERROR: GET request to retrieve friends failed"
        echo "  expected: $EXPECTED_RESPONSE"
        echo "       got: $ACTUAL_RESPONSE"
        exit 1
    fi

    # Test DELETE /friends/:id
    ACTUAL_RESPONSE=$(curl --silent --request DELETE \
        --write-out "%{http_code}\n" \
        --output /dev/null \
        http://localhost:8080/friends/1
    )

    EXPECTED_RESPONSE='200'

    compare_responses $ACTUAL_RESPONSE $EXPECTED_RESPONSE "DELETE request to remove a friend"

    # Test GET /birthdays
    TODAYS_DATE=$(date '+%d/%m/%Y')

    ACTUAL_RESPONSE=$(curl --silent --request POST \
        --write-out "%{http_code}\n" \
        --output /dev/null \
        --data "{\"Name\":\"Jack Reacher\",\"LastContacted\":\"06/06/2023\",\"Birthday\":\"$TODAYS_DATE\"}" \
        http://localhost:8080/friends)

    EXPECTED_RESPONSE='201'

    # Checking to make sure the friend addition worked properly
        if [[ $ACTUAL_RESPONSE != $EXPECTED_RESPONSE ]]; then
        echo "ERROR: "POST request to add a friend failed"
        echo "  expected: $EXPECTED_RESPONSE"
        echo "       got: $ACTUAL_RESPONSE"
        exit 1
    else
        echo "INFO: "POST request to add a friend passed"
    fi

    ACTUAL_RESPONSE=$(curl --silent \
        --write-out "%{http_code}\n" \
        --output /dev/null \
        http://localhost:8080/birthdays
    )

    EXPECTED_RESPONSE='200'

        # Checking to make sure the friend addition worked properly
        if [[ $ACTUAL_RESPONSE != $EXPECTED_RESPONSE ]]; then
        echo "ERROR: "GET request to get todays birthdays failed"
        echo "  expected: $EXPECTED_RESPONSE"
        echo "       got: $ACTUAL_RESPONSE"
        exit 1
    else
        echo "INFO: "GET request to get todays birthdays passed"
    fi

}

trap cleanup EXIT

build_and_run
main
