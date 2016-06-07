#!/bin/sh

source 0-setup.sh

etId=$1

cat > tmp <<foo
{
    "title": "High Severity",
    "condition": {
        "eventtype_ids": ["$etId"],
        "query": {
            "query": {
                "bool": {
                    "must": [
                        { "match": {"severity": 5} },
                        { "match": {"code": "PHONE"} }
                    ]
                }
            }
        }
     },
    "job": { "task": "alert the user" }
}
foo

json=`cat tmp`

echo
echo POST /v2/trigger
echo "$json"

ret=`curl -S -s -XPOST -d "$json" $WHOST/v2/trigger`

echo RETURN:
echo $ret
echo
