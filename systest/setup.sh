#!/bin/bash
set -e

export THEDOMAIN=int.geointservices.io
export PZSERVER=pz-gateway.$THEDOMAIN

export PZKEY=`cat ~/.pzkey | grep $PZSERVER | cut -f 2 -d ":" | cut -d \" -f 2`

curl="curl -S -s -u $PZKEY: -H Content-Type:application/json"

url="http://$PZSERVER"
workflowurl="http://pz-workflow.$THEDOMAIN"

extract() {
    item=$1
    str=$2

    echo "$str" | grep $item | cut -f 2 -d ":" | cut -d \" -f 2
}

#curl -S -s -u $PZUSER:$PZPASS -H Content-Type:application/json https://pz-gateway.int.geointservices.io/key

