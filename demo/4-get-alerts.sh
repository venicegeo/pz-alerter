#!/bin/sh

source 0-setup.sh

echo
echo GET /alerts

ret=`curl -S -s -XGET -d "$json" $WHOST/v1/alerts`

echo RETURN:
echo $ret
echo
