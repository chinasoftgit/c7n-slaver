FROM golang:1.9.4-alpine3.7 as builder
WORKDIR /go/src/github.com/choerodon/c7n-slaver
ADD . .

RUN go build ./cmd/slaver

FROM alpine

COPY --from=builder /go/src/github.com/choerodon/c7n-slaver/slaver /c7n-slaver

RUN \
set -ex \
   && apk add --no-cache ca-certificates

EXPOSE 9000 9001
CMD ["/c7n-slaver"]