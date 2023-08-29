FROM golang:1.19.2-alpine as builder

RUN apk add bash

RUN apk add --no-cache openssh-client ansible git

WORKDIR /workspace
COPY . ./

RUN go build -o myapp ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /workspace .

RUN chmod a+x myapp
CMD ["./myapp"]
