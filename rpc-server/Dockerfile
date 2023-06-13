FROM golang:1.18
WORKDIR /app
COPY . .
RUN go build -o main
EXPOSE 8888
CMD ["./main"]