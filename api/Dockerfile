FROM golang:latest AS build

WORKDIR /go/src/app
COPY . /go/src/app

RUN go build -o /go/bin/app

FROM gcr.io/distroless/base-debian11
COPY --from=build /go/bin/app /
CMD ["/app"]

