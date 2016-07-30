#!/bin/bash

source setup.sh

eventTypeId=$1
if [ "$eventTypeId" == "" ]
then
    echo "error: \$eventTypeId missing"
    exit 1
fi

#echo GET /eventType/$eventTypeId

ret=$($curl -XGET "$url"/eventType/$eventTypeId)

#echo RETURN:
#echo "$ret"

echo `extract eventTypeId "$ret"`
