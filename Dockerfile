#
# build stage
#
FROM golang:1.20-buster AS build

WORKDIR /

COPY go.mod go.sum *.go Makefile ./
COPY internal/ ./internal/
RUN go mod download

RUN BINARY_OUTPUT=/forwarder OS=linux ARCH=amd64 make build
RUN chgrp 0 /forwarder && chmod g+rwX /forwarder

#
# runtime
#
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=build /forwarder /forwarder

CMD ["/forwarder"]
