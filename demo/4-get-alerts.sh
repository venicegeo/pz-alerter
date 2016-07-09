#!/bin/bash

# shellcheck disable=SC1091
source 0-setup.sh

echo
echo GET /alert

ret=$(curl -S -s -XGET "$WHOST"/alert)

echo RETURN:
echo "$ret"
echo
