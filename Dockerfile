FROM golang:1.18-bullseye
COPY . .
RUN CGO_ENABLED=0 go build -o /exporter

FROM scratch
COPY --from=0 /exporter /exporter
ENTRYPOINT ["/exporter"]
