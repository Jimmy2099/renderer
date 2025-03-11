#!/bin/bash

IMAGE_NAME="renderer_thirdparty"

docker build -t "$IMAGE_NAME" .
echo "Please follow the license terms. Do not distribute this image without proper authorization."
echo "If you need to push to a personal registry, use the following command:"
echo "docker tag $IMAGE_NAME your_registry/your_repo:$IMAGE_NAME"
echo "docker push your_registry/your_repo:$IMAGE_NAME"
echo "Done."

