FROM golang:1.24-alpine AS build

RUN apk add --no-cache ca-certificates

WORKDIR /app

# cache module downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=${VERSION}" -a -o server ./cmd/server


FROM alpine:3.21

RUN apk --no-cache add ca-certificates && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app/
COPY --from=build /app/server .
COPY --from=build /app/config/*.yml ./config/
RUN chown -R app:app /app

USER app

CMD ["./server"]
