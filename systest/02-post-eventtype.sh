#!/bin/bash

source setup.sh

name=testevent-`date "+%s"`

json='{
    "name": "'"$name"'",
    "mapping": {
        "alpha": "string",
        "beta": "integer"
    }
}'

#echo POST /eventType
#echo "$json"

ret=$($curl -d "$json" "$url"/eventType)

#echo RETURN:
#echo "$ret"

echo `extract eventTypeId "$ret"`
