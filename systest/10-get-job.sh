#!/bin/bash
set -e

source setup.sh

jobId=$1
[ "$jobId" != "" ] || ( echo error: \$jobId missing ; exit 1 )

echo GET /job/$jobId
ret=$($curl -X GET $url/job/$jobId)

#echo RETURN:
#echo "$ret"

echo `extract jobId "$ret"`
