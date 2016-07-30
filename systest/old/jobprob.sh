#!/bin/bash

url="http://pz-workflow.int.geointservices.io"

eventTypeId="77bbe4c6-b1ac-4bbb-8e86-0f6e6a731c39"
serviceId="61985d9c-d4d0-45d9-a655-7dcf2dc08fad"

json='{
    "name": "systest$jobprob",
    "enabled": true,
    "condition": {
        "eventTypeIds": ["'"$eventTypeId"'"],
        "query": {
            "query": {
                "match": {"beta": 17}
            }
        }
    },
    "job": {
        "createdBy": "test",
        "jobType": {
            "type": "execute-service",
            "data": {
                "dataInputs": {
                    "": {
                        "content": {"name":"ME", "count":"5"},
                        "type":     "body",
                        "mimeType": "application/json"
                    }
                },
                "dataOutput": [
                    {
                        "mimeType": "application/json",
                        "type":     "text"
                    }
                ],
                "serviceId": "'"$serviceId"'"
            }
        }
    }
}'

echo
echo POST /trigger
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$url"/trigger)

echo RETURN:
echo "$ret"
echo
