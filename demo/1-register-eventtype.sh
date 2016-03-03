#!/bin/sh

json=`cat <<foo
{
    "name": "USData",
    "mapping": {
        "itemId":   "string",
        "severity": "integer",
        "problem":  "string"
    }
}
foo`

echo POST:
echo "$json"

ret=`curl -S -s -XPOST -d "$json" http://pz-workflow.cf.piazzageo.io/v1/eventtypes`

echo RETURN:
echo $ret
