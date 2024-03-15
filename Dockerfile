FROM alpine:latest
USER root
COPY s1mple /usr/local/bin
RUN mkdir /opt/s1mple && chmod +x /usr/local/bin/s1mple
COPY doc_go_template.txt /opt/s1mple
ENTRYPOINT ["s1mple"]