FROM golang:1.25.1-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

# Copy dependency project
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/main.go

FROM alpine:3.22

WORKDIR /app
RUN apk add --no-cache make

# Copy the built binary
COPY --from=builder /build/server ./server

# Copy the Makefile
COPY --from=builder /build/Makefile ./Makefile

# Copy migrations & seeds
COPY --from=builder /build/db/migrations ./db/migrations
COPY --from=builder /build/db/seeds ./db/seeds

RUN chmod +x server

EXPOSE 3000

CMD [ "./server"]