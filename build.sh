#!/bin/bash

# set default environment (if no argument passed, default to pre, support prod, pre)
MODE="pre"

# set default version (if no argument passed, default to latest)
VERSION="latest"

MAKE_MODE="debug"
REPOSITORY="mytemplate"
IMAGE="mytemplate"

# set docker account
USER_NAME="*"
USER_PASSWD="*"

# parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --mode=*)
            MODE="${1#*=}"
            shift
            ;;
        --version=*)
            VERSION="${1#*=}"
            shift
            ;;
        *)
            echo "❌ Unknown option: $1"
            echo "Usage: ./build.sh [--mode=prod|pre] [--version=xxx]"
            exit 1
            ;;
    esac
done

# check arguments
if [ "$MODE" != "prod" ] && [ "$MODE" != "pre" ]; then
    echo "❌  invalid argument: $MODE 'prod' or 'pre'"
    exit 1
fi

if [ "$MODE" == "prod" ]; then
    MAKE_MODE=""
fi

# ensure buildx is usable
docker buildx inspect --bootstrap > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "⚙️  Creating buildx builder..."
    docker buildx create --use
fi

# build image

IMAGE_NAME="${IMAGE}:${VERSION}"
echo "🔨 start building $IMAGE_NAME..."

DOCKERFILE="./cmd/test/Dockerfile"
if [ "$MODE" == "prod" ]; then
    DOCKERFILE="./cmd/prod/Dockerfile"
fi

if ! docker buildx build \
    --platform linux/amd64 \
    --network=host \
    --no-cache \
    -f $DOCKERFILE \
    --build-arg BUILD_MODE=$MAKE_MODE \
    -t $IMAGE_NAME \
    . \
    --load; then
    echo "❌ image build failed"
    exit 1
fi
echo "✅ image built successfully: $MAKE_MODE"

# login image registry
echo "🔐 login image registry..."
if ! docker login -u $USER_NAME -p $USER_PASSWD; then
    echo "❌ login failed"
    exit 1
fi


echo "📤 start pushing $IMAGE_NAME..."
docker tag $IMAGE:$VERSION $REPOSITORY/$IMAGE:$VERSION
if ! docker push $REPOSITORY/$IMAGE_NAME; then
    echo "❌ image push failed"
    docker logout
    exit 1
fi

# build complete
echo "✅ images built and pushed successfully: $IMAGE_NAME"