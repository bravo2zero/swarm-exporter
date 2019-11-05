FROM golang:1.13

RUN mkdir -p /app
COPY . /app
WORKDIR /app

RUN go mod download && go build -o exporter
# default metrics port
EXPOSE 8080

CMD [ "/app/exporter" ]
