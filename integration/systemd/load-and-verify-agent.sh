#!/bin/bash
# Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# 
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
# 
#     http://aws.amazon.com/apache2.0/
# 
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

# This wrapper script ensures that we've loaded the desired ECS agent
# and have attached the "latest" tag to it.
# USAGE: load-and-verify-agent.sh DEFAULT_TAG

docker ping > /dev/null || {
    echo "Cannot talk to Docker daemon. Aborting." >&2
    exit 1
}

DEFAULT_TAG=$1 ; shift
AGENT_IMAGE_NAME=amazon/amazon-ecs-agent
DESIRED_IMAGE_FILE=/var/cache/ecs/desired-image

if [ -f "$DESIRED_IMAGE_FILE" ]; then
    DESIRED_TAG=$(head -n1 "$DESIRED_IMAGE_FILE")
else
    DESIRED_TAG="$DEFAULT_TAG"
fi

IMAGE_NAME="${AGENT_IMAGE_NAME}:${DESIRED_TAG}"


image_id=$(docker inspect "$IMAGE_NAME")
latest_image_id=$(docker inspect "${AGENT_IMAGE_NAME}:latest")

if -z "$image_id" ; then
    # We already have the desired image. Ensure that "latest" points to it.
    docker tag ${AGENT_IMAGE_NAME}:${image_id##sha256:} ${AGENT_IMAGE_NAME}:latest
    
