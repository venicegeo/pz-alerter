#!/bin/bash

PZDOMAIN=int.geointservices.io

url="https://pz-logger.$PZDOMAIN/admin/stats"

echo
echo GET $url
echo "$json"

ret=$(curl -S -s -XGET -d "$json" "$url")

echo RETURN:
echo "$ret"
echo
