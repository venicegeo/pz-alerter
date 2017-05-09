#!/bin/bash
INDEX_NAME=eventtypes005
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

EventTypeMapping='
	"EventType": {
		"dynamic": "strict",
		"properties": {
			"eventTypeId": {
				"type": "keyword"
			},
			"name": {
				"type": "keyword"
			},
			"createdOn": {
				"type": "date",
				"format": "yyyy-MM-dd'\''T'\''HH:mm:ssZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSSSZZ"
			},
			"createdBy": {
				"type": "keyword"
			},
			"mapping": {
				"dynamic": "false",
				"type": "object"
			}
		}
	}'
IndexSettings="
{
	"\""mappings"\"": {
		$EventTypeMapping
	}
}"


bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP "$IndexSettings" "$EventTypeMapping" $TESTING
