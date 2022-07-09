FROM golang:1.18
RUN mkdir /app
WORKDIR /app
COPY . .
RUN go build -o main .
EXPOSE 8081
CMD ["/app/main"]
