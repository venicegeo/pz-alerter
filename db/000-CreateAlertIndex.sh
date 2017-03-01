#!/bin/bash
INDEX_NAME=alerts003
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

AlertMapping='
	"Alert": {
		"properties": {
			"alertId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"triggerId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"jobId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"eventId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"createdBy": {
				"type": "string",
				"index": "not_analyzed"
			},
			"createdOn": {
				"type": "date"
			}
		}
	}'

IndexSettings="
{
	"\""mappings"\"": {
		$AlertMapping
	}
}"

echo $IndexSettings >> db/index.txt
echo $AlertMapping >> db/mapping.txt



bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP $TESTING