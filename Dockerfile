FROM golang:1.20

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o /aws-web-server

CMD [ "/aws-web-server" ]