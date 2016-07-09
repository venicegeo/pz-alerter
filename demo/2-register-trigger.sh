#!/bin/bash

# shellcheck disable=SC1091
source 0-setup.sh

etId=$1

cat > tmp <<foo
{
    "title": "High Severity",
    "condition": {
        "eventTypeIds": ["$etId"],
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
                "serviceId": "ddd5134"
            }
        }
    }
}
foo

json=$(cat tmp)

echo
echo POST /v2/trigger
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$WHOST"/trigger)

echo RETURN:
echo "$ret"
echo

rm tmp
