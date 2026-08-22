# build
FROM golang:1.26-alpine AS builder

WORKDIR /usr/src/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o /out/shortener ./cmd/shortener

# runtime
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/shortener /shortener

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT [ "/shortener" ]
