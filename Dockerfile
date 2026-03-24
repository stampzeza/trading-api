FROM golang:1.25

WORKDIR /app

# copy mod ก่อน
COPY go.mod go.sum ./

RUN go mod download

# copy code
COPY . .

RUN go build -o app

CMD ["./app"]