#!/bin/sh
#set -x

curl -XPOST -d '{
    "type": "USDataFound",
	"date": "22 Jan",
	"data": {}
	}' http://localhost:12342/events
echo

#curl -XGET http://localhost:12342/events
