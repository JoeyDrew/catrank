FROM golang:latest
WORKDIR /app

# This can be split into multiple COPY commands to make use of better caching
COPY catrank.go go.mod go.sum cats.json style.css ./
COPY templates/ ./templates/
RUN go build -o catrank .
EXPOSE 8080
CMD ["/app/catrank"]
