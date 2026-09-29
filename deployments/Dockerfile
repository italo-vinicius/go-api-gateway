FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/go-api-gateway ./cmd/go-api-gateway
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/mockservice ./cmd/mockservice
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/go-api-gateway /go-api-gateway
EXPOSE 8080 9090
USER nonroot:nonroot
ENTRYPOINT ["/go-api-gateway"]
CMD ["serve", "--config", "/etc/go-api-gateway/config.yaml"]

FROM gcr.io/distroless/static-debian12:nonroot AS mockservice
COPY --from=build /out/mockservice /mockservice
USER nonroot:nonroot
ENTRYPOINT ["/mockservice"]
