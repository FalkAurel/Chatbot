FROM alpine:latest

RUN apk add go lshw
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /app
RUN mkdir data
RUN touch data/default_prompt.txt
COPY . .

EXPOSE 8080
RUN go build main.go

CMD ./main

