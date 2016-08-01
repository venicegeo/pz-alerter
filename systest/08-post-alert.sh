#!/bin/bash

source setup.sh

json='{
    "triggerId": "22",
    "eventId": "33",
    "jobId": "44"
}'

#echo POST /alert
#echo "$json"

ret=$($curl -XPOST -d "$json" "$workflowurl"/alert)

#echo RETURN:
#echo "$ret"

echo `extract alertId "$ret"`
