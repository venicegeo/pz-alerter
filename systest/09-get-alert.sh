#!/bin/bash

source setup.sh

alertId=$1
if [ "$alertId" == "" ]
then
    echo "error: \$alertId missing"
    exit 1
fi

#echo GET /alert/$alertId

ret=$($curl -XGET "$url"/alert/$alertId)

#echo RETURN:
#echo "$ret"

echo `extract alertId "$ret"`
