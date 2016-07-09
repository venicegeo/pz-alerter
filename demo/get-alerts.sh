#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

echo
echo GET /alert

ret=$(curl -S -s -XGET "$url"/alert)

echo RETURN:
echo "$ret"
echo
