#!/bin/bash

source setup.sh

triggerId=$1
if [ "$triggerId" == "" ]
then
    echo "error: \$triggerId missing"
    exit 1
fi

#echo GET /trigger

ret=$($curl -S -s -XGET -d "$json" "$workflowurl"/trigger/$triggerId)

#echo RETURN:
#echo "$ret"

echo `extract triggerId "$ret"`
