FROM mycode-run

RUN apk add --no-cache go && mkdir -p /go/src

ENV GOPATH=/go/
