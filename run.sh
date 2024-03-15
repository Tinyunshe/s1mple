#!/bin/bash
image=$1
img_dir="/opt/s1mple"
docker container rm -f s1mple
docker run -itd --name s1mple -p 8081:8080 -v ./s1mple_config_prod.yaml:$img_dir/s1mple_config.yaml $image --config $img_dir/s1mple_config.yaml