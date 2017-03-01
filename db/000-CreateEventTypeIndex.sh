#!/bin/bash
INDEX_NAME=eventtypes003
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

EventTypeMapping='
	"EventType": {
		"properties": {
			"eventTypeId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"name": {
				"type": "string",
				"index": "not_analyzed"
			},
			"createdOn": {
				"type": "date"
			},
			"createdBy": {
				"type": "string",
				"index": "not_analyzed"
			},
			"mapping": {
				"dynamic": false,
				"properties": {}
			}
		}
	}'
IndexSettings="
{
	"\""mappings"\"": {
		$EventTypeMapping
	}
}"
echo $IndexSettings >> db/index.txt
echo $EventTypeMapping >> db/mapping.txt



bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP $TESTING