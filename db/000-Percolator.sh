#!/bin/bash
INDEX_NAME=perc001
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3

PercMapping='
	"queries": {
		"dynamic": "strict",
		"properties": {
			"query": {
				"type": "percolator"
			}
		}
	},
	"doctype": {
	}'
IndexSettings="
{
	"\""settings"\"": {
		"\""index.mapping.coerce"\"": false
	},
	"\""mappings"\"": {
		$PercMapping
	}
}"


bash db/CreateIndex.sh $INDEX_NAME $ALIAS_NAME $ES_IP "$IndexSettings" "$PercMapping" $TESTING
