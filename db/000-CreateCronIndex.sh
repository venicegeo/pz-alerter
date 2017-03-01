#!/bin/bash
INDEX_NAME=crons003
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

CronMapping='
	"Cron": {
		"properties": {
			"eventTypeId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"eventId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"data": {
				"properties": {}
			},
			"createdBy": {
				"type": "string",
				"index": "not_analyzed"
			},
			"createdOn": {
				"type": "date"
			},
			"cronSchedule": {
				"type": "string",
				"index": "not_analyzed"
			}
		}
	}'
IndexSettings="
{
	"\""settings"\"": {
		"\""index.mapping.coerce"\"": false
	},
	"\""mappings"\"": {
		$CronMapping
	}
}"

echo $IndexSettings >> db/index.txt
echo $CronMapping >> db/mapping.txt



bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP $TESTING