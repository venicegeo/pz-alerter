#!/bin/bash

# optional
msg="Yow. $1"

PZDOMAIN=int.geointservices.io

# %z gives "-0700" but we need "-07:00"
d=`date +%Y-%m-%dT%T`
#d=2006-01-02T15:04:05+07:00
d=$d-05:00

input='{
    "service":  "alpha",
    "address":  "1.2.3.4",
    "createdOn": "'"$d"'",
    "severity": "Debug",
    "message":  "'"$msg"'"
}'

url="https://pz-logger.$PZDOMAIN/message"
echo
echo POST $url
echo "$input"

echo RETURN:
curl -S -s -XPOST -d "$input" "$url"
echo
