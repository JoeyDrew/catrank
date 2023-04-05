FROM golang:latest
WORKDIR /app
COPY . .
RUN go build -buildvcs=false -o catrank . 
EXPOSE 8080
CMD ["/app/catrank"]