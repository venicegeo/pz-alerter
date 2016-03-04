#!/bin/sh

source 0-setup.sh

curl -S -s -XGET $WHOST/v1/eventtypes
echo
curl -S -s -XGET $WHOST/v1/events/USData
echo
curl -S -s -XGET $WHOST/v1/triggers
echo
curl -S -s -XGET $WHOST/v1/alerts
echo

for i in 1 2 3 4 5 6 7 8 9 10
do
    curl -S -s -XDELETE  $WHOST/v1/eventtypes/ET$i
    echo
    curl -S -s -XDELETE  $WHOST/v1/events/USData/E$i
    echo
    curl -S -s -XDELETE  $WHOST/v1/triggers/TRG$i
    echo
    curl -S -s -XDELETE  $WHOST/v1/alerts/A$i
    echo
done

curl -S -s -XGET $WHOST/v1/eventtypes
echo
curl -S -s -XGET $WHOST/v1/events/USData
echo
curl -S -s -XGET $WHOST/v1/triggers
echo
curl -S -s -XGET $WHOST/v1/alerts
echo
