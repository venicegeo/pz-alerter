#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

echo
echo GET /alert

ret=$(curl -S -s -XGET "$url"/alert?perPage=10&sortBy=createdOn&triggerId=1705dce3-324b-4d76-a125-ffc3a7cb0016)

echo RETURN:
echo "$ret"
echo
