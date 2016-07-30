#!/bin/bash

source setup.sh

eventId=$1
if [ "$eventId" == "" ]
then
    echo "error: \$eventId missing"
    exit 1
fi

#echo GET /event/$eventId

ret=$($curl -XGET "$url"/event/$eventId)

#echo RETURN:
#echo "$ret"

echo `extract eventId "$ret"`
