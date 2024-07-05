FROM golang:1.22.5-alpine3.20
RUN apk add git
WORKDIR /app

COPY . .
RUN go mod download
RUN go build -buildvcs=false -o /pgrest

EXPOSE 8080

CMD [ "/pgrest" ]
