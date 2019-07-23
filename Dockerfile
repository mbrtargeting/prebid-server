FROM alpine:3.8 AS build
WORKDIR /go/src/github.com/prebid/prebid-server/
RUN apk add -U --no-cache go git dep musl-dev
ENV GOPATH /go
ENV CGO_ENABLED 0
COPY ./ ./
RUN dep ensure
RUN go build .

FROM alpine:3.8 AS release
MAINTAINER Hans Hjort <hans.hjort@xandr.com>
WORKDIR /usr/local/bin/
COPY --from=build /go/src/github.com/prebid/prebid-server/prebid-server .
COPY run.sh .
COPY pbs.json .
COPY static static/
COPY stored_requests/data stored_requests/data
RUN apk add -U --no-cache ca-certificates mtr
RUN apk add --no-cache bash
RUN apk add --no-cache curl
EXPOSE 8003
EXPOSE 8080

#ENTRYPOINT ["/usr/local/bin/prebid-server"]
#CMD ["-v", "1", "-logtostderr"]

ENTRYPOINT ["/usr/local/bin/run.sh"]
