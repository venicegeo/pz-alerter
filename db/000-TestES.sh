#!/bin/bash
INDEX_NAME=testelasticsearch004
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

TestElasticsearchMapping='
	"TestElasticsearch":{
		"properties":{
			"id": {
				"type":"string"
			},
			"data": {
				"type":"string"
			},
			"tags": {
				"type":"string"
			},
			"value": {
				"type": "long"
			}
		}
	}'
IndexSettings="
{
	"\""mappings"\"": {
		$TestElasticsearchMapping
	}
}"


bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP "$IndexSettings" "$TestElasticsearchMapping" $TESTING
