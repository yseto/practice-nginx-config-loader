FROM golang:latest AS build

WORKDIR /go/src/app
COPY . /go/src/app

RUN go build -o /go/bin/app

#FROM gcr.io/distroless/base-debian11
FROM nginx:latest
COPY --from=build /go/bin/app /
COPY config.template /usr/local/etc/config.template
CMD ["/app", \
    "-template", "/usr/local/etc/config.template", \
    "-backend", "http://api:8080/api.json", \
    "-output", "/etc/nginx/conf.d/defaults.conf", \
    "--", "nginx", "-g", "daemon off;"]
