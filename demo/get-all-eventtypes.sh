#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

echo
echo GET /event

ret=$(curl -S -s -XGET "$url"/eventType?perPage=5)

echo RETURN:
echo "$ret"
echo
