FROM golang:1.18-bullseye
RUN go build -ldflags="-extldflags=-static" -o /exporter

FROM scratch
COPY --from=0 /exporter /exporter
ENTRYPOINT ["/exporter"]
