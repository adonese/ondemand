from golang:latest

COPY . /app
WORKDIR /app
RUN go build
EXPOSE 6662
CMD ["/app/ondemand"]
