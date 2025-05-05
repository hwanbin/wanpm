FROM golang:1.23.1-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o app ./cmd/api

FROM gcr.io/distroless/base-debian12

WORKDIR /bin
COPY --from=builder /build/app /bin/app
ENTRYPOINT [ "/bin/app" ]