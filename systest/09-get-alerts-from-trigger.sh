#!/bin/bash

source setup.sh

triggerId=$1
if [ "$triggerId" == "" ]
then
    echo "error: \$triggerId missing"
    exit 1
fi

echo
echo GET /alert

ret=$($curl -XGET $url/alert/triggerId=$triggerId)

echo RETURN:
echo "$ret"
echo
