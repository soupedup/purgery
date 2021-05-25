FROM golang:1.16-alpine as build
RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY go.* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build \
    -mod readonly \
    -o binary \
    .

FROM scratch
COPY --from=build /build/binary /
CMD [ "/binary" ]
