#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

eventtypeId=$1
serviceId=$2
[ "$eventTypeId" != "" ] || ( echo error: \$eventTypeId missing ; exit 1 )
[ "$serviceId" != "" ] || ( echo error: \$serviceId missing ; exit 1 )

json='{
    "title": "High Severity",
    "condition": {
        "eventTypeIds": ["'"$eventTypeId"'"],
        "query": {
            "query": {
                "bool": {
                    "must": [
                        { "match": {"severity": 5} },
                        { "match": {"code": "PHONE"} }
                    ]
                }
            }
        }
    },
    "job": {
        "task": "alert the user",
        "userName": "test",
        "jobType": {
            "type": "execute-service",
            "data": {
                "serviceId": "'"$serviceId"'"
            }
        }
    }
}'

echo
echo POST /v2/trigger
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url"/trigger)

echo RETURN:
echo "$ret"
echo
