# Build container
#FROM alpine:latest
FROM golang:1.26.1-alpine
RUN apk -U upgrade --no-cache && apk add --no-cache git && rm -rf /var/cache/apk/* && mkdir /build
WORKDIR /build
RUN git clone https://github.com/johannes-kuhfuss/stt-service.git
WORKDIR /build/stt-service
RUN go build -o /build/stt-service/stt-service /build/stt-service/main.go
# Run container
FROM alpine:3.23.3
RUN apk -U upgrade --no-cache && rm -rf /var/cache/apk/* && mkdir /app
WORKDIR /app
COPY --from=0 /build/stt-service/stt-service /app/stt-service
COPY --from=0 /build/stt-service/templates /app/templates
COPY --from=0 /build/stt-service/bootstrap /app/bootstrap
RUN addgroup -g 10000 servicegroup && adduser -s /sbin/nologin -G servicegroup -D -H -u 10000 serviceuser
USER serviceuser
ENV STT_PATH=/uploads
HEALTHCHECK --interval=120s --timeout=5s CMD wget -q --spider http://localhost:8080/ || exit 1
ENTRYPOINT ["/app/stt-service"]
EXPOSE 8080/tcp