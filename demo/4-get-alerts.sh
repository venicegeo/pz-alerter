#!/bin/sh

source 0-setup.sh

echo
echo GET /v2/alert

ret=`curl -S -s -XGET -d "$json" $WHOST/v2/alert`

echo RETURN:
echo $ret
echo
