FROM golang:1.20.2-bullseye
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o exporter

FROM scratch
COPY --from=0 /app/exporter /exporter
ENTRYPOINT ["/exporter"]
