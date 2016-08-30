#!/bin/bash

source setup.sh

eventTypeId=$1
if [ "$eventTypeId" == "" ]
then
    echo "error: \$eventTypeId missing"
    exit 1
fi

#echo GET /eventType/$eventTypeId

ret=$($curl -XDELETE "$url"/eventType/$eventTypeId)

#echo RETURN:
#echo "$ret"

echo `extract eventTypeResponse "$ret"`
