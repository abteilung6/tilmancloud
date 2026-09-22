FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN GOPROXY=https://proxy.golang.org,direct go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /manager ./cmd/operator

FROM gcr.io/distroless/static-debian12:nonroot@sha256:cdf4daaf154e3e27cfffc799c16f343a384228f38646928a1513d925f473cb46
COPY --from=build /manager /manager
USER 65532:65532
ENTRYPOINT ["/manager"]
