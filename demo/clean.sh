#!/bin/sh

source 0-setup.sh

curl -S -s -XGET $WHOST/v1/eventtypes
curl -S -s -XGET $WHOST/v1/events/USData
curl -S -s -XGET $WHOST/v1/triggers
curl -S -s -XGET $WHOST/v1/alerts

for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30
do
    curl -S -s -XDELETE  $WHOST/v1/eventtypes/ET$i
    curl -S -s -XDELETE  $WHOST/v1/events/USData/E$i
    curl -S -s -XDELETE  $WHOST/v1/triggers/TRG$i
    curl -S -s -XDELETE  $WHOST/v1/alerts/A$i
done

curl -S -s -XGET $WHOST/v1/eventtypes
curl -S -s -XGET $WHOST/v1/events/USData
curl -S -s -XGET $WHOST/v1/triggers
curl -S -s -XGET $WHOST/v1/alerts
