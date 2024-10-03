FROM golang:1.23.1 AS build

WORKDIR /go/src/app
COPY . .

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go mod download
RUN templ generate

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM scratch

COPY --from=build /go/bin/app /
EXPOSE 8080
CMD ["/app"]