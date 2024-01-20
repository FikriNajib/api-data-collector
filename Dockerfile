FROM golang:alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .
RUN go mod tidy

RUN go build -o data-collector-api

EXPOSE ${PORT}

CMD [ "./data-collector-api", "serve" ]