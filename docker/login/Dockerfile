FROM golang AS builder
WORKDIR $GOPATH/src/github.com/jadefish/avatar
COPY . .

# RUN go get -u github.com/golang/dep/...
# RUN dep ensure -vendor-only

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o login ./cmd/login

FROM golang:1-alpine
WORKDIR /app
COPY --from=builder /go/src/github.com/jadefish/avatar/login /app/login
CMD ["/app/login"]
