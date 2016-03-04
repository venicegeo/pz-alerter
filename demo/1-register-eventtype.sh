#!/bin/sh

source 0-setup.sh

cat > tmp <<foo
{
    "name": "USData",
    "mapping": {
        "itemId":   "string",
        "severity": "integer",
        "problem":  "string"
    }
}
foo

json=`cat tmp`

echo POST:
echo "$json"

ret=`curl -S -s -XPOST -d "$json" $WHOST/v1/eventtypes`

echo RETURN:
echo $ret
