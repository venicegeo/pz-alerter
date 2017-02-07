#!/bin/sh

export THEDOMAIN=int.geointservices.io
export PZSERVER=piazza.$THEDOMAIN
export PZKEY=`cat ~/.pzkey | jq -r .'"'$PZSERVER'"'`
curl="curl -S -s -u $PZKEY: -H Content-Type:application/json"
url="http://pz-logger.$THEDOMAIN"

json='
{
	"Foo": "bar"
}'

#echo $json

out=`$curl -X POST -d "$json" $url/query`

echo $out
