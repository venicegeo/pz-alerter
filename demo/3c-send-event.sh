#!/bin/sh

source 0-setup.sh

etId=$1

cat > tmp <<foo
{
    "eventtype_id": "$etId",
    "date": "2007-05-05T14:30:00Z",
    "data": {
        "filename": "dataset-c",
        "severity": 5,
        "code": "PHONE"
    }
}
foo

json=`cat tmp`

echo
echo POST /events/USDataEvent
echo "$json"

ret=`curl -S -s -XPOST -d "$json" $WHOST/v1/events/USDataEvent`

echo RETURN:
echo $ret
echo
