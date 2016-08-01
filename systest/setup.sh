#!/bin/bash
set -e

space="stage"
space="int"
domain=geointservices.io

export PZKEY=`cat ~/.pzkey | grep $space | cut -f 2 -d ":" | cut -d \" -f 2`
#echo $PZKEY
export PZDOMAIN=$space.$domain

curl="curl -S -s -u $PZKEY: -H Content-Type:application/json"

url="http://pz-gateway.$PZDOMAIN"
workflowurl="http://pz-workflow.$PZDOMAIN"

extract() {
    item=$1
    str=$2

    echo "$str" | grep $item | cut -f 2 -d ":" | cut -d \" -f 2
}

#curl -S -s -u $PZUSER:$PZPASS -H Content-Type:application/json https://pz-gateway.int.geointservices.io/key

