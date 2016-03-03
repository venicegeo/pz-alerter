#!/bin/sh

etId=$1

cat > tmp <<foo
{
    "type": "$etId",
    "date": "2007-05-05T14:30:00Z",
    "data": {
        "itemId": "e38ad2", 
        "severity": 4,
        "problem": "us-phone"
    }
}
foo

json=`cat tmp`

echo POST /events/$etId
echo "$json"

ret=`curl -S -s -XPOST -d "$json" http://pz-workflow.cf.piazzageo.io/v1/events/$etId`

echo RETURN:
echo $ret
