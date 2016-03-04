#!/bin/sh

source 0-setup.sh

etId=$1

#query='{\"query\" : {\"bool\": {\"must\": [{\"match\" : {\"severity\" : 4}},{\"match\" : {\"problem\" : \"us-bbox\"}}]}}}'
#echo $query


cat > tmp <<foo
{
    "title": "my found-a-bad-telephone-number trigger",
    "condition": {
        "type": "$etId",
        "query": {
            "query": {
                "match": {
                    "severity": 2
                }
            }
        },
        "job": "do the thing!"
    }
}
foo

json=`cat tmp`

echo POST /triggers
echo "$json"

ret=`curl -S -s -XPOST -d "$json" $WHOST/v1/triggers`

echo RETURN:
echo $ret
