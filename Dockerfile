########################
# build container
FROM golang:1.19.2-bullseye AS build
RUN apt update && apt install -y git make && rm -rf /var/lib/apt/lists/*

## build the binary
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build

########################
# runtime container
FROM debian:bullseye-slim
RUN apt update && apt install -y ca-certificates && rm -rf /var/lib/apt/lists/*

## create a user
RUN useradd -m -s /bin/bash -u 3939 slashes
USER slashes

## copy the binary
COPY --from=build /build/bin/slashes /usr/local/bin/slashes

## run the binary
ENTRYPOINT ["/usr/local/bin/slashes"]
