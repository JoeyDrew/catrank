FROM golang:latest
WORKDIR /app
COPY . .
RUN go build -buildvcs=false -o catrank . 
EXPOSE 8000
CMD ["/app/catrank"]