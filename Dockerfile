FROM golang:1.26.3-alpine3.23 AS build

COPY . /src/
RUN cd /src && go build -o bin main.go

FROM alpine:3.23.4

COPY --from=build /src/config.json /config.json
COPY --from=build /src/events /events
COPY --from=build /src/bin /bin

ENTRYPOINT [ "bin" ]
