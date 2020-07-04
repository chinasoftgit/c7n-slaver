FROM golang:1.9.4-alpine3.7@sha256:d17e39f48a1ba5c2bb97f1198a370253d1e67a42c87bd1056d864dd6b0a3f463 
WORKDIR /go/src/github.com/choerodon/c7n-slaver
ADD . .

RUN go build ./cmd/slaver

