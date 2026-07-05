FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wealth-lens-api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates \
    && addgroup -S wealthlens \
    && adduser -S -G wealthlens wealthlens
COPY --from=build /out/wealth-lens-api /usr/local/bin/wealth-lens-api

USER wealthlens
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/wealth-lens-api"]
