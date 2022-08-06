FROM golang:alpine

WORKDIR /app/
COPY . ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o pravdabot

COPY ./docker/docker-entrypoint.sh .
RUN chmod +x /app/docker-entrypoint.sh
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/pravdabot"]