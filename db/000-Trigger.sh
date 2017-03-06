#!/bin/bash
INDEX_NAME=triggers004
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
			"name": {
				"type": "string",
				"index": "not_analyzed"
			},
			"createdOn": {
				"type": "date",
				"format": "yyyy-MM-dd'\''T'\''HH:mm:ssZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSSSZZ"
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
				"type": "boolean"
			},
			"condition": {
				"dynamic": "false",
				"type": "object"
			},
			"job": {
				"properties": {
					"createdBy": {
						"type": "string",
						"index": "not_analyzed"
					},
					"jobType": {
						"dynamic": "false",
						"type": "object"
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


bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP "$IndexSettings" "$TriggerMapping" $TESTING
