# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.23.4 AS build

WORKDIR /src
# Copy the Go Modules manifests
COPY go.mod go.sum ./
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

# Build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/manager .

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/manager /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
