#!/bin/bash
# set default environment (if no argument passed, default to pre, support prod, pre)
MODE="pre"

# set default version (if no argument passed, default to latest)
VERSION="latest"

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
            echo "📘 Usage: ./build.sh [--mode=prod|pre] [--version=xxx]"
            exit 1
            ;;
    esac
done

# check arguments
if [ "$MODE" != "prod" ] && [ "$MODE" != "pre" ]; then
    echo "❌  invalid argument: $MODE 'prod' or 'pre'"
    exit 1
fi

IMAGE_HEAD="$REPOSITORY/$IMAGE"
IMAGE_NAME="${IMAGE_HEAD}:${VERSION}"
echo "✅ Starting deployment $IMAGE_NAME..."

# login image registry
echo "🔐 login image registry..."
if ! docker login -u $USER_NAME -p $USER_PASSWD; then
    echo "❌ login failed"
    exit 1
fi

echo "🛠️   Pulling image: $IMAGE_NAME..."
IMAGE_NAME=$IMAGE_NAME docker compose pull

echo "🛑 Stopping and removing containers..."
IMAGE_NAME=$IMAGE_NAME docker compose down

echo "🚀 Starting containers..."
IMAGE_NAME=$IMAGE_NAME docker compose up -d

echo "✅ Docker Compose started with image: $IMAGE_NAME."