FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/eyex ./cmd/api

FROM scratch
WORKDIR /app
COPY --from=build /out/eyex /app/eyex
COPY frontend/html-js /app/frontend/html-js
COPY openapi.yaml /app/openapi.yaml
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/eyex"]
