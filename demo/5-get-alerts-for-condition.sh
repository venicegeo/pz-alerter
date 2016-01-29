#!/bin/sh
#set -x

id=$1

#echo $id

curl -XGET http://localhost:12342/alerts/bycondition/$id
echo
