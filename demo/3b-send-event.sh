#!/bin/bash

# shellcheck disable=SC1091
source 0-setup.sh

etId=$1

cat > tmp <<foo
{
    "eventtype_id": "$etId",
    "date": "2007-05-05T14:30:00Z",
    "data": {
        "filename": "dataset-b",
        "severity": 2,
        "code": "PHONE"
    }
}
foo

json=$(cat tmp)

echo
echo POST /v2/event
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$WHOST"/v2/event)

echo RETURN:
echo "$ret"
echo

rm tmp
