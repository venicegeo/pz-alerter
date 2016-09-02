#!/bin/bash

source setup.sh

eventId=$1
if [ "$eventId" == "" ]
then
    echo "error: \$eventId missing"
    exit 1
fi

#echo DELETE /event/$eventId

ret=$($curl -XDELETE "$url"/event/$eventId)

#echo RETURN:
#echo "$ret"

echo `extract eventResponse "$ret"`
