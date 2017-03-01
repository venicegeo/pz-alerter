#!/bin/bash
INDEX_NAME=events004
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

EventMapping='
	"_default_": {
		"dynamic": "false",
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
		"\""index.mapping.coerce"\"": false,
		"\""index.version.created"\"": 2010299
	},
	"\""mappings"\"": {
		$EventMapping
	}
}"

echo $IndexSettings >> db/index.txt
echo $EventMapping >> db/mapping.txt



bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP $TESTING