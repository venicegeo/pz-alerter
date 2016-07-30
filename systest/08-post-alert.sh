#!/bin/bash

source setup.sh
url="http://pz-workflow.$PZDOMAIN"

json='{
    "triggerId": "22",
    "eventId": "33",
    "jobId": "44"
}'

#echo POST /alert
#echo "$json"

ret=$($curl -XPOST -d "$json" "$url"/alert)

#echo RETURN:
#echo "$ret"

echo `extract alertId "$ret"`
