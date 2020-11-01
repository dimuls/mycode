FROM alpine

ARG NSJAIL_VERSION=3.0

RUN apk add --no-cache build-base libstdc++ bison bsd-compat-headers flex \
linux-headers protobuf-dev libnl-dev libnl3-dev git \
&& git clone --depth=1 --branch=${NSJAIL_VERSION} https://github.com/google/nsjail.git /nsjail \
&& cd /nsjail \
&& make \
&& mv /nsjail/nsjail /usr/bin/nsjail \
&& rm -rf /nsjail \
&& apk del --purge build-base bison bsd-compat-headers flex linux-headers

RUN apk add time --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing

COPY mycode-run /usr/bin/mycode-run
