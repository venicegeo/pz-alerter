#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

echo
echo GET /event

ret=$(curl -S -s -XGET "$url"/eventType?perPage=13&order=desc&page=1)

echo "$ret" | grep createdOn

echo "$ret"
