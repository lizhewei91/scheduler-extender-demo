FROM golang:1.22-alpine as build

WORKDIR /go/src/scheduler-extender-demo
COPY . .
RUN go build -o /go/bin/scheduler-extender-demo cmd/scheduler-extender/main.go

FROM alpine

COPY --from=build /go/bin/scheduler-extender-demo /usr/bin/scheduler-extender-demo

CMD ["scheduler-extender-demo"]