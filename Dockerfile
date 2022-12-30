FROM golang:1.18 as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
# Workaround for Private Github REPOs
ARG GITHUB_TOKEN
RUN go env -w GOPRIVATE="*"
RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/function ./
FROM alpine:3.17
COPY --from=0 /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]