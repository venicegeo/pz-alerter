#!/bin/bash

set -e

source setup.sh

serviceId=`sh 01-register-service.sh`
echo 01 ServiceId: $serviceId

eventTypeId=`sh 02-post-eventtype.sh`
echo 02 EventTypeId: $eventTypeId

t=`sh 03-get-eventtype.sh $eventTypeId`
echo . 03 check: $t

triggerId=`sh 04-post-trigger.sh $eventTypeId $serviceId`
echo 04 TriggerId: $triggerId

t=`sh 05-get-trigger.sh $triggerId`
echo . 05 check: $t

eventIdY=`sh 06-post-event-yes.sh $eventTypeId`
echo 06 EventIdY: $eventIdY

eventIdN=`sh 06-post-event-no.sh $eventTypeId`
echo 06 EventIdN: $eventIdN

t=`sh 07-get-event.sh $eventIdY`
echo . 07 check: $t

alertId=`sh 08-post-alert.sh`
echo . 08 AlertId: $alertId

t=`sh 09-get-alert.sh $alertId`
echo . 09 check: $t

alertId=`sh 09-get-alert-from-trigger.sh $triggerId`
echo 09 AlertId: $alertId

jobId=`sh 09-get-job-from-alert.sh $alertId`
echo 09 JobId: $jobId

jobId=`sh 10-get-job.sh $jobId`
echo . 10 JobId: $jobId

dataId=`sh 10-get-data-from-job.sh $jobId`
echo 10 DataId: $dataId

info=`sh 11-get-data.sh $dataId`
echo 11 results: $info
