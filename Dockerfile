FROM golang:1.19

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o /aws-web-server

EXPOSE 80

CMD [ "/aws-web-server" ]