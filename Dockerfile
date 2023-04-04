FROM golang:latest
WORKDIR /app
COPY . .
RUN go build -o catrank .
EXPOSE 8080
CMD ["/app/catrank"]