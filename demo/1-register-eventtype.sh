#!/bin/bash

# shellcheck disable=SC1091
source 0-setup.sh

cat > tmp <<foo
{
    "eventTypeId": "1234",
    "name": "USDadtaEvent",
    "mapping": {
        "filename": "string",
        "code":     "string",
        "severity": "integer"
    }
}
foo

json=$(cat tmp)

echo
echo POST /eventType
echo "$json"

ret=$(curl -S -s -XPOST -d "$json" "$WHOST"/eventType)

echo RETURN:
echo "$ret"
echo

rm tmp
