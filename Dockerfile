# Build static binaries, then ship them on a minimal base image.
FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/cli ./cmd/cli

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
# Admin CLI (setpass, bootstrap). The server has no Go toolchain, so the only
# way to run it in production is from this image: --entrypoint /cli
COPY --from=build /out/cli /cli
EXPOSE 8080
ENTRYPOINT ["/api"]
