#!/bin/bash
set -o errexit

BRANCH_NAME=$(git branch --show-current)
IMAGE_NAME="kalmonipa/howarethey:$BRANCH_NAME"
CONTAINER_NAME="$BRANCH_NAME"

## Clean up
cleanup() {
    if [ "$(docker ps -a | grep $CONTAINER_NAME)" ]; then
        echo "INFO: Stopping $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" > /dev/null
        echo "INFO: Removing $CONTAINER_NAME"
        docker rm "$CONTAINER_NAME" > /dev/null
   fi
}

# TODO: Get this bit working from within the Go tests so I can remove this script
build_image() {
    echo "Building image $IMAGE_NAME"

    ## Build the image to test
    docker build . --tag "$IMAGE_NAME"

    sleep 5
}

trap cleanup EXIT

# Skip building the image to save time running tests
if [[ "$BUILD_IMAGE" != "false" ]]; then
    build_image
else
    echo "INFO: Skipping image build"
fi

go test -v ./pkg/test/integration_test
