FROM golang:1.9.4-alpine3.7@sha256:d17e39f48a1ba5c2bb97f1198a370253d1e67a42c87bd1056d864dd6b0a3f463 as builder
WORKDIR /go/src/github.com/choerodon/c7n-slaver
ADD . .

RUN go build ./cmd/slaver

FROM alpine@sha256:3b3f647d2d99cac772ed64c4791e5d9b750dd5fe0b25db653ec4976f7b72837c

COPY --from=builder /go/src/github.com/choerodon/c7n-slaver/slaver /c7n-slaver

RUN \
set -ex \
   && apk add --no-cache ca-certificates

EXPOSE 9000 9001
CMD ["/c7n-slaver"]
