# Build container
#FROM alpine:latest
FROM golang:1.26.1-alpine
RUN apk -U upgrade --no-cache && apk add --no-cache git && rm -rf /var/cache/apk/* && mkdir /build
WORKDIR /build
RUN git clone https://github.com/johannes-kuhfuss/xcode-service.git
WORKDIR /build/xcode-service
RUN go build -o /build/xcode-service/xcode-service /build/xcode-service/main.go
# Run container
FROM alpine:3.23.1
RUN apk -U upgrade --no-cache && apk add --no-cache ffmpeg && rm -rf /var/cache/apk/* && mkdir /app
WORKDIR /app
COPY --from=0 /build/xcode-service/xcode-service /app/xcode-service
COPY --from=0 /build/xcode-service/templates /app/templates
COPY --from=0 /build/xcode-service/bootstrap /app/bootstrap
RUN addgroup -g 101 servicegroup && adduser -s /sbin/nologin -G servicegroup -D -H -u 101 serviceuser
USER serviceuser
ENV XCODE_PATH=/uploads
HEALTHCHECK --interval=120s --timeout=5s CMD wget -q --spider http://localhost:8080/ || exit 1
ENTRYPOINT ["/app/xcode-service"]
EXPOSE 8080/tcp