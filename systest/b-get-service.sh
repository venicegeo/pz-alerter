#!/bin/bash
set -e

source setup.sh

serviceId=$1
[ "$serviceId" != "" ] || ( echo error: \$serviceId missing ; exit 1 )

ret=$($curl -X GET $url/service/$serviceId)

echo "$ret"
