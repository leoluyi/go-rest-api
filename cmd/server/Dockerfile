FROM registry.suse.com/bci/golang:1.24 AS build

WORKDIR /app

# cache module downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=${VERSION}" -a -o server ./cmd/server


FROM registry.suse.com/bci/bci-micro:latest

WORKDIR /app/
COPY --from=build --chown=65532:65532 /app/server .
COPY --from=build --chown=65532:65532 /app/config/*.yml ./config/

USER 65532:65532

CMD ["./server"]
