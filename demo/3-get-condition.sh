#!/bin/sh

#set -x

id=$1

curl -XGET http://localhost:12342/conditions/$id

echo
