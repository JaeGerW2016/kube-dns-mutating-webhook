FROM golang:1.12 as builder

ENV GOPROXY https://goproxy.io
ENV GO111MODULE on

WORKDIR /go/cache
COPY [ "go.mod","go.sum","./"]
RUN go mod download

WORKDIR /go/release
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kube-dns-mutating-webhook-server .

FROM 314315960/alpine-base:3.9
COPY --from=builder /go/release/kube-dns-mutating-webhook-server kube-dns-mutating-webhook-server
ENTRYPOINT ["./kube-dns-mutating-webhook-server"]

