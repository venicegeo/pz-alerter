#!/bin/bash

source setup.sh

triggerId=$1
if [ "$triggerId" == "" ]
then
    echo "error: \$triggerId missing"
    exit 1
fi

#echo DELETE /trigger/$triggerId

ret=$($curl -XDELETE "$url"/trigger/$triggerId)

#echo RETURN:
#echo "$ret"

echo `extract triggerResponse "$ret"`
