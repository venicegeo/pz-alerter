#!/bin/bash

# shellcheck disable=SC1091
source 0-setup.sh

echo
echo GET /v2/alert

ret=$(curl -S -s -XGET "$WHOST"/v2/alert)

echo RETURN:
echo "$ret"
echo
