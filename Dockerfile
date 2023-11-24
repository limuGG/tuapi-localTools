FROM golang:1.20-alpine as builder
WORKDIR /opt
COPY . .
RUN go build -trimpath -ldflags "-s -w" -o tools.out . && chmod +x tools.out

FROM alpine:latest
WORKDIR /opt
COPY --from=builder /opt/tools.out .