#!/bin/sh

source 0-setup.sh

etId=$1

cat > tmp <<foo
{
    "type": "$etId",
    "date": "2007-05-05T14:30:00Z",
    "data": {
        "itemId": "eb872", 
        "severity": 2,
        "problem": "us-bbox"
    }
}
foo

json=`cat tmp`

echo POST /events/USData
echo "$json"

ret=`curl -S -s -XPOST -d "$json" $WHOST/v1/events/USData`

echo RETURN:
echo $ret
