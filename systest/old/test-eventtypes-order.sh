#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

echo
echo GET /event

#ret=$(curl -S -s -XGET "$url"/eventType?perPage=13&order=asc&page=1&sortBy=createdOn)

ret=$(curl -S -s -XGET "$url"/eventType\?order=asc\&perPage=13\&sortBy=eventTypeId)

echo "$ret" | grep eventTypeId

#echo "$ret"
