#!/bin/bash

source setup.sh

eventTypeId="77bbe4c6-b1ac-4bbb-8e86-0f6e6a731c39"
serviceId="61985d9c-d4d0-45d9-a655-7dcf2dc08fad"

args='{\"name\":\"ME\", \"count\":5}'
echo $args

json='{
    "type": "execute-service",
    "data": {
        "dataInputs": {
            "": {
                "content": "'"$args"'",
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
}'

# run the job
echo POST /job
echo "$json"
ret=$($curl -XPOST -d "$json" "$url"/job)

# show results
echo RETURN:
echo "$ret"

# get returned job id
jobId=`extract jobId "$ret"`
echo JOB ID: $jobId

# check the job status
ret=$($curl -XGET $url/job/$jobId)
echo "$ret"

# check the job status harder
sleep 2
ret=$($curl -XGET $url/job/$jobId)
echo "$ret"

# get the data id
dataId=`extract dataId "$ret"`
echo DataId: $dataId

# get the data
ret=$($curl -XGET $url/data/$dataId)
echo "$ret"
