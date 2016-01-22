#!/bin/sh
#set -x

curl -XPOST -d '{
    "title": "test for US data",
    "type": "USDataFound",
	"user_id": "mpg",
	"start_date": "4:31pm"
	}' http://localhost:12342/conditions
echo

#curl -XGET http://localhost:12342/conditions
