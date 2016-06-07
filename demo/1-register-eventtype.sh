#!/bin/sh

source 0-setup.sh

cat > tmp <<foo
{
    "name": "USDataEvent",
    "mapping": {
        "filename": "string",
        "code":     "string",
        "severity": "integer"
    }
}
foo

json=`cat tmp`

echo
echo POST /v2/eventType
echo "$json"

ret=`curl -S -s -XPOST -d "$json" $WHOST/v2/eventType`

echo RETURN:
echo $ret
echo
