#!/bin/bash
INDEX_NAME=events006
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

EventMapping='
	"_default_": {
		"dynamic": "strict",
		"properties": {
			"eventTypeId": {
				"type": "keyword"
			},
			"eventId": {
				"type": "keyword"
			},
			"data": {
				"dynamic": "true",
				"type": "object"
			},
			"createdBy": {
				"type": "keyword"
			},
			"createdOn": {
				"type": "date",
				"format": "yyyy-MM-dd'\''T'\''HH:mm:ssZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSSZZ||yyyy-MM-dd'\''T'\''HH:mm:ss.SSSSSSSZZ"
			},
			"cronSchedule": {
				"type": "keyword"
			}
		}
	}'
IndexSettings="
{
	"\""settings"\"": {
		"\""index.mapping.coerce"\"": false
	},
	"\""mappings"\"": {
		$EventMapping
	}
}"


bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP "$IndexSettings" "$EventMapping" $TESTING
