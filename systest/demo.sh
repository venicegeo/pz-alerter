#!/bin/bash

serviceId=`sh 01-register-service.sh`
echo ServiceId: $serviceId

eventTypeId=`sh 02-post-eventtype.sh`
echo EventTypeId: $eventTypeId

t=`sh 03-get-eventtype.sh $eventTypeId`
echo check: $t

triggerId=`sh 04-post-trigger.sh $eventTypeId $serviceId`
echo TriggerId: $triggerId

t=`sh 05-get-trigger.sh $triggerId`
echo check: $t

eventIdY=`sh 06-post-event-yes.sh $eventTypeId`
echo EventIdY: $eventIdY

eventIdN=`sh 06-post-event-no.sh $eventTypeId`
echo EventIdN: $eventIdN

t=`sh 07-get-event.sh $eventIdY`
echo check: $t

alertId=`sh 08-post-alert.sh`
echo AlertId: $alertId

t=`sh 09-get-alert.sh $alertId`
echo check: $t

alertIds=`sh 09-get-alerts-from-trigger.sh $triggerId`
echo AlertIds: $alertIds

#jobId=`sh 09-get-job-from-alert.sh $alertId`
#echo JobId: $jobId
