#!/bin/bash
INDEX_NAME=testelasticsearch003
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
			}
		}
	}'
IndexSettings="
{
	"\""mappings"\"": {
		$TestElasticsearchMapping
	}
}"

echo $IndexSettings >> db/index.txt
echo $TestElasticsearchMapping >> db/mapping.txt



bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP $TESTING