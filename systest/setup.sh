#!/bin/bash
set -e

curl="curl -S -s -u $PZKEY: -H Content-Type:application/json"

url="http://pz-gateway.$PZDOMAIN"

extract() {
    item=$1
    str=$2

    echo "$str" | grep $item | cut -f 2 -d ":" | cut -d \" -f 2
}
