# syntax=docker/dockerfile:1

FROM golang:1.22.2

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY auth ./auth
COPY config ./config
COPY db ./db
COPY docs ./docs
COPY handlers ./handlers
COPY models ./models
COPY services ./services
COPY setup ./setup
COPY utils ./utils
COPY webhooks ./webhooks
COPY data ./data
COPY web-ui ./web-ui
COPY config.yaml key.pem cert.pem main.go rsakey.pem build.sh run.sh ./

RUN CGO_ENABLED=1 GOOS=linux go build -o dingus-server

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 443

# Run
CMD ["/app/run.sh"]