FROM FROM arm64v8/golang:latest
WORKDIR /go/src/github.com/choerodon/c7n-slaver
ADD . .

RUN go build ./cmd/slaver

