FROM golang
WORKDIR $GOPATH/src/github.com/jadefish/avatar
COPY . .

# RUN go get -u github.com/golang/dep/...
# RUN dep ensure -vendor-only

RUN go build -a -o login ./cmd/login
CMD ["./login"]
