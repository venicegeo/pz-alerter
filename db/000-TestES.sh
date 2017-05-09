#!/bin/bash
INDEX_NAME=testelasticsearch005
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

TestElasticsearchMapping='
	"TestElasticsearch":{
		"dynamic": "strict",
		"properties":{
			"id": {
				"type":"keyword"
			},
			"data": {
				"type":"keyword"
			},
			"tags": {
				"type":"keyword"
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