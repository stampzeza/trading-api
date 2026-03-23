FROM golang:1.22

WORKDIR /app

# copy go mod ก่อน (เร็วขึ้น + fix error)
COPY go.mod go.sum ./
RUN go mod download

# copy code
COPY . .

RUN go build -o app

CMD ["./app"]