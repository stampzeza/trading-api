# ใช้ Go image
FROM golang:1.22

# ตั้ง working directory
WORKDIR /app

# copy file
COPY . .

# download dependency
RUN go mod tidy

# build binary
RUN go build -o app

# run app
CMD ["./app"]