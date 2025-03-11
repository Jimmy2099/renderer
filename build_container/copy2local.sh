#!/bin/bash

IMAGE_NAME="renderer_thirdparty"

echo "Please pull and tag your image locally first:"
echo "  docker pull your_registry/your_repo:$IMAGE_NAME"
echo "  docker tag your_registry/your_repo:$IMAGE_NAME $IMAGE_NAME"
echo ""
docker cp $IMAGE_NAME:/renender .
echo "Done."
