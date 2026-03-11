FROM nexus-dev-image.fulltrust.link/base-images/golang:1.24-alpine AS build
WORKDIR /temp/src
ADD . /temp/src

COPY go.mod go.sum ./
COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main main.go

# run image no need golang, so use alpine as base image
FROM nexus-dev-image.fulltrust.link/base-images/alpine:latest
WORKDIR /app

ENV ENV=pre

COPY --from=build /temp/src/main /app
COPY --from=build /temp/src/etc /app/etc


# expose port for external access
EXPOSE 8080
# internal service
# EXPOSE 8901
# metric
EXPOSE 8081
CMD ["/app/main"]