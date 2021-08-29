FROM golang:1.17-alpine as build
RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY go.* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build \
    -mod readonly \
    -o binary \
    .

FROM bash
COPY --from=build /build/binary /usr/local/bin/purgery
COPY start.sh /start.sh
CMD [ "/start.sh" ]
