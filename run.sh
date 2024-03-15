#!/bin/bash
image=$1
docker container rm -f s1mple
docker run -itd --name s1mple -p 8081:8080 -v ./s1mple_config_prod.yaml:/opt/s1mple/s1mple_config.yaml -v ./doc_go_template.txt:/opt/s1mple/doc_go_template.txt $image --config /opt/s1mple_config.yaml