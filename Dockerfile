ARG GO_VERSION
ARG OS_CODENAME
ARG BASEIMAGE
FROM golang:${GO_VERSION}-${OS_CODENAME} AS builder

WORKDIR /actions-runner-cleaner
COPY . .
ARG OUTPUT_DIR
ENV CGO_ENABLED=0
RUN make build

FROM ${BASEIMAGE}

WORKDIR /
ARG OUTPUT_DIR
COPY --from=builder /actions-runner-cleaner/${OUTPUT_DIR}/actions-runner-cleaner .
ENTRYPOINT ["/actions-runner-cleaner"]
