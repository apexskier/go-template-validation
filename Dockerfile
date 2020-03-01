FROM golang:1.13-buster as build
WORKDIR /app
ADD . /app
RUN go test ./...
RUN go build -o /binary

FROM gcr.io/distroless/base-debian10
COPY --from=build /binary /
COPY --from=build /app/index.html /
CMD ["/binary"]
