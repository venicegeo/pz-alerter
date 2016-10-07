#!/bin/bash

PZDOMAIN=int.geointservices.io

url="https://pz-logger.$PZDOMAIN/message?perPage=3"

echo
echo GET $url
echo "$json"

ret=$(curl -S -s -XGET -d "$json" "$url")

echo RETURN:
echo "$ret"
echo
