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
echo POST $WHOST/v1/eventtypes
echo "$json"

ret=`curl -S -s -XPOST -d "$json" $WHOST/v1/eventtypes`

echo RETURN:
echo $ret
echo
