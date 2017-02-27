#!/bin/bash
INDEX_NAME=events004
ALIAS_NAME=$1
ES_IP=$2
TESTING=$3
aliases=_aliases
IndexSettings='
{
	"settings": {
		"index.mapping.coerce": false,
		"index.version.created": 2010299
	},
	"mappings": {
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
		}
	}
}
	'

if [[ $ALIAS_NAME == "" ]]; then
  echo "Please specify an alias name as argument 1"
  exit 1
fi

if [[ $ES_IP == "" ]]; then
  echo "Please specify the elasticsearch ip as argument 2"
  exit 1
fi 

if [[ $ES_IP != */ ]]; then
  ES_IP="$ES_IP/"
fi

if [[ $TESTING == "" ]]; then
  $TESTING="false"
fi

function removeAliases {
  echo "Running remove alias function"
  crash=$1

  #
  # Search for indices that are using the alias we are trying to set
  #

  getAliasesCurl=`curl -XGET -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$cat/aliases" --write-out %{http_code} 2>/dev/null`
  http_code=`echo $catCurl | cut -d] -f2`
  if [[ "$http_code" != 200 ]]; then
    echo "Status code $http_code returned from catting aliases"
    if [ "$crash" == true ]; then
      exit 1
    fi
  fi

  #
  # Extract index names that are using the alias from the above result
  #

  regex=""\""alias"\"":"\""$ALIAS_NAME"\"","\""index"\"":"\""([^"\""]+)"
  temp=`echo $getAliasesCurl|grep -Eo $regex | cut -d\" -f8`
  indexArr=(${temp// / })
  echo "Found ${#indexArr[@]} indices currently using alias $ALIAS_NAME: ${indexArr[@]}"

  #
  # Remove alias from all above indices
  #

  for index in ${indexArr[@]}
  do
    removeAliasCurl=`curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
      "\""actions"\"" : [
          { "\""remove"\"" : { "\""index"\"" : "\""$index"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } }
      ]
    }" "$ES_IP$aliases" --write-out %{http_code} 2>/dev/null`
    http_code=`echo $catCurl | cut -d] -f2`
    if [[ $removeAliasCurl != '{"acknowledged":true}200' ]]; then
      echo "Failed to remove alias $ALIAS_NAME on index $index. Code: $http_code"
      if [ "$crash" == true ]; then    
        exit 1
      fi
    fi
    echo "Removed alias $ALIAS_NAME on index $index"
  done
}

function createAlias {
  echo "Running create alias function"
  crash=$1

  #
  # Create alias on our index
  #

  createAliasCurl=`curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
      "\""actions"\"" : [
          { "\""add"\"" : { "\""index"\"" : "\""$INDEX_NAME"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } }
      ]
  }" "$ES_IP$aliases" --write-out %{http_code} 2>/dev/null`
  http_code=`echo $catCurl | cut -d] -f2`
  if [[ $createAliasCurl != '{"acknowledged":true}200' ]]; then
    echo "Failed to create alias $ALIAS_NAME on index $INDEX_NAME. Code: $http_code"
    if [ "$crash" == true ]; then
      exit 1
    fi
  fi
  echo "Created alias $ALIAS_NAME on index $INDEX_NAME"
}

#
# Check to see if index already exists
#

echo "Checking to see if index $INDEX_NAME already exists..."
cat=_cat
catCurl=`curl -X GET -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$cat/indices" --write-out %{http_code} 2>/dev/null`
http_code=`echo $catCurl | cut -d] -f2`
if [[ "$http_code" != 200 ]]; then
  echo "Status code $http_code returned while checking indices"
  exit 1
fi

if [[ $catCurl == *""\""index"\"":"\""$INDEX_NAME"\"""* ]]; then
  removeAliases false
  createAlias true
  echo "Index already exists"
exit 0
fi

#
# Create the index
#

echo "Creating index $INDEX_NAME with mappings..."
createIndexCurl=`curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "$IndexSettings" "$ES_IP$INDEX_NAME" --write-out %{http_code} 2>/dev/null`
echo $createIndexCurl
http_code=`echo $catCurl | cut -d] -f2`
if [[ $createIndexCurl != '{"acknowledged":true}200' ]]; then
  echo "Failed to create index $INDEX_NAME. Code: $http_code"
  exit 1
fi

#
# If testing, create two indices that have the alias we are trying to set
#

if [ "$TESTING" = true ] ; then
    echo "Creating test indices..."
    orange=orange
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{}" "$ES_IP$orange" --write-out %{http_code}; echo " "
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [ { "\""add"\"" : { "\""index"\"" : "\""orange"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } } ]
    }" "$ES_IP$aliases" --write-out %{http_code}; echo " "
    cherry=cherry
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{}" "$ES_IP$cherry" --write-out %{http_code}; echo " "
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [ { "\""add"\"" : { "\""index"\"" : "\""cherry"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } } ]
    }" "$ES_IP$aliases" --write-out %{http_code}; echo " "
fi

removeAliases false

createAlias true

echo 
echo "Success!"
