
FROM golang:1.22


WORKDIR /


COPY go.mod go.sum ./

RUN go mod download

COPY . ./


RUN go build -o /graduation ./cmd/server/main.go

EXPOSE 8088

CMD ["/graduation"]