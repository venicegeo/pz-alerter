#!/bin/sh

etId=$1

cat > t <<foo
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

json=`cat t`

echo POST /events/$etId
echo "$json"

ret=`curl -S -s -XPOST -d "$json" http://pz-workflow.cf.piazzageo.io/v1/events/$etId`

echo RETURN:
echo $ret
