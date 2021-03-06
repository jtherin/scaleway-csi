FROM golang:1.14-alpine as builder

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /go/src/github.com/scaleway/scaleway-csi

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/ cmd/
COPY scaleway/ scaleway/
COPY driver/ driver/

ARG TAG
ARG COMMIT_SHA
ARG BUILD_DATE
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -ldflags "-w -s -X github.com/scaleway/scaleway-csi/driver.driverVersion=${TAG} -X github.com/scaleway/scaleway-csi/driver.buildDate=${BUILD_DATE} -X github.com/scaleway/scaleway-csi/driver.gitCommit=${COMMIT_SHA} " -o scaleway-csi ./cmd/scaleway-csi

FROM scratch
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/scaleway/scaleway-csi/scaleway-csi .
ENTRYPOINT ["/scaleway-csi"]
