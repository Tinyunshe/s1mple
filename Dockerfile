# FROM alpine:latest
# USER root
# COPY s1mple /usr/local/bin
# RUN mkdir /opt/s1mple && chmod +x /usr/local/bin/s1mple
# COPY doc_go_template.txt /opt/s1mple
# ENTRYPOINT ["s1mple"]



FROM docker-mirrors.alauda.cn/library/golang:1.21 as builder

WORKDIR /workspace

COPY . .

RUN ls && pwd

RUN cd cmd && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../s1mple && cd ..

# FROM docker-mirrors.alauda.cn/library/alpine:latest
FROM build-harbor.alauda.cn/ops/alpine:3.19

COPY --from=builder  /workspace/s1mple /usr/local/bin
RUN mkdir /opt/s1mple && chmod +x /usr/local/bin/s1mple
COPY --from=builder /workspace/doc_go_template.txt /opt/s1mple/doc_go_template.txt
ENTRYPOINT ["s1mple"]