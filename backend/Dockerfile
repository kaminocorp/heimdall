FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/heimdall ./cmd/heimdall

FROM alpine:3.21
COPY --from=build /bin/heimdall /usr/local/bin/heimdall
EXPOSE 8080
ENTRYPOINT ["heimdall"]
