#!/bin/bash

DOMAIN=int.geointservices.io

url="https://pz-uuidgen.$DOMAIN/uuids?count=3"
echo
echo POST $url
echo "$json"

ret=$(curl -S -s -XPOST "$url")

echo RETURN:
echo "$ret"
echo
