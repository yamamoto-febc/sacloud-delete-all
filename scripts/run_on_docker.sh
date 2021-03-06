#!/bin/bash

# 注:ポート番号は固定
docker run -it --rm \
  --name $1 \
  -e SAKURACLOUD_ACCESS_TOKEN \
  -e SAKURACLOUD_ACCESS_TOKEN_SECRET \
  -e SAKURACLOUD_ZONES \
  -e SAKURACLOUD_TRACE_MODE \
  $1:latest ${@:2}