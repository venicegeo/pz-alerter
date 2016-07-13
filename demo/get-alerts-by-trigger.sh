#!/bin/bash

url="http://pz-workflow.$PZDOMAIN"

triggerId=$1
if [ "$triggerId" == "" ]
then
    echo "error: \$triggerId missing"
    exit 1
fi

echo
echo GET /alert

cmd="curl -S -s -XGET $url/alert?perPage=10&sortBy=createdOn&triggerId=$triggerId"
echo $cmd

ret=$($cmd)

echo RETURN:
echo "$ret"
echo
