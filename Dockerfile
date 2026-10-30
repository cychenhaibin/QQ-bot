FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/qq-bot ./cmd/bot

FROM alpine:3.22
RUN addgroup -S qqbot && adduser -S -G qqbot qqbot
COPY --from=build /out/qq-bot /usr/local/bin/qq-bot
USER qqbot
ENTRYPOINT ["/usr/local/bin/qq-bot"]
