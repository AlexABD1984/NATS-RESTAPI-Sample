FROM golang:latest 
WORKDIR /go/src/app
RUN go get -v -u github.com/xeipuuv/gojsonschema
RUN go get -v -u github.com/nats-io/go-nats
RUN go get -v -u github.com/gorilla/mux
WORKDIR /go/src/app
COPY . .
RUN go build -o main .
RUN ls
CMD ["/go/src/app/main"]
