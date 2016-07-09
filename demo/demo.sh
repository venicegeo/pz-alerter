#!/bin/bash

sh post-service.sh > tmp.1
serviceId=`grep serviceId tmp.1 | cut -f 2 -d ":" | cut -d \" -f 2`
echo ServiceId: $serviceId

sh post-eventtype.sh > tmp.2
eventTypeId=`grep eventTypeId tmp.2 | cut -f 2 -d ":" | cut -f 1 -d "," | cut -d \" -f 2`
echo EventTypeId: $eventTypeId

sh post-trigger.sh $eventTypeId $serviceId > tmp.3
triggerId=`grep triggerId tmp.3 | cut -f 2 -d ":" | cut -f 1 -d "," | cut -d \" -f 2`
echo TriggerId: $triggerId

sh post-event-no.sh $eventTypeId > tmp.4
eventNoId=`grep eventId tmp.4 | cut -f 2 -d ":" | cut -f 1 -d "," | cut -d \" -f 2`
echo EventId/no: $eventNoId

sh post-event-yes.sh $eventTypeId > tmp.5
eventYesId=`grep eventId tmp.5 | cut -f 2 -d ":" | cut -f 1 -d "," | cut -d \" -f 2`
echo EventId/yes: $eventYesId

sleep 3
sh get-all-alerts.sh > tmp.6
grep $eventYesId tmp.6

