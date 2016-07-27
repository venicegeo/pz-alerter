#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

echo
echo GET /alert

ret=$(curl -S -s -XGET "$url"/alert?perPage=10&sortBy=createdOn)

echo RETURN:
echo "$ret"
echo
