#!/bin/bash

source setup.sh

triggerId=$1
if [ "$triggerId" == "" ]
then
    echo "error: \$triggerId missing"
    exit 1
fi

json='{
    "enabled": false
}'


#echo PUT /trigger
#echo "$json"

ret=$($curl -S -s -XPUT -d "$json" "$url"/trigger/$triggerId)

#echo RETURN:
#echo "$ret"

echo `extract triggerId "$ret"`
