#!/bin/bash
INDEX_NAME=triggers003
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

TriggerMapping='
	"Trigger": {
		"properties": {
			"triggerId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"title": {
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
			"eventTypeId": {
				"type": "string",
				"index": "not_analyzed"
			},
			"enabled": {
				"type": "boolean",
				"index": "not_analyzed"
			},
			"condition": {
				"dynamic": true,
				"properties": {}
			},
			"job": {
				"properties": {
					"createdBy": {
						"type": "string",
						"index": "not_analyzed"
					},
					"jobType": {
						"dynamic": true,
						"properties": {}
					}
				}
			},
			"percolationId": {
				"type": "string",
				"index": "not_analyzed"
			}
		}
	}'
IndexSettings="
{
	"\""mappings"\"": {
		$TriggerMapping
	}
}"

echo $IndexSettings >> db/index.txt
echo $TriggerMapping >> db/mapping.txt



bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP $TESTING