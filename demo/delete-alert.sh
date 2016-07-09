#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

alertId=$1
if [ "$alertId" == "" ]
then
    echo "error: \$alertId missing"
    exit 1
fi

echo
echo DELETE "$url"/alert/"$alertId"
echo "$json"

ret=$(curl -S -s -XDELETE -d "$json" "$url"/alert/"$alertId")

echo RETURN:
echo "$ret"
echo
