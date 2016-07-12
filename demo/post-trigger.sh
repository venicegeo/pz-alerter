#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

eventTypeId=$1
if [ "$eventTypeId" == "" ]
then
    echo "error: \$eventTypeId missing"
    exit 1
fi

serviceId=$2
if [ "$serviceId" == "" ]
then
    echo "error: \$serviceId missing"
    exit 1
fi

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
        "createdBy": "test",
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
