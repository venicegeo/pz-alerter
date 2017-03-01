INDEX_NAME=$1
ALIAS_NAME=$2
ES_IP=$3
IndexSettings=$4
MappingEsc=$5
TESTING=$6

MappingEsc=${MappingEsc//"\""/'\"'}

aliases=_aliases
cat=_cat

function failure {
	echo "{"\""status"\"":"\""failure"\"","\""message"\"":"\""$1"\""}"
	exit 1
}

function success {
	echo "{"\""status"\"":"\""success"\"","\""message"\"":"\""$1"\"","\""mapping"\"":"\""{$MappingEsc}"\""}"
	exit 0
}

function printIfTesting {
  if [ "$TESTING" = true ] ; then
    echo "$1"
  fi
}

if [[ $ALIAS_NAME == "" ]]; then
  failure "Please specify an alias name as argument 1"
fi

if [[ $ES_IP == "" ]]; then
  failure "Please specify the elasticsearch ip as argument 2"
fi 

if [[ $ES_IP != */ ]]; then
  ES_IP="$ES_IP/"
fi

if [[ $TESTING == "" ]]; then
  TESTING=false
fi

function tryCrash {
	if [ "$1" == true ] ; then
	  failure "$2"
	fi
}

function removeAliases {
  printIfTesting "Running remove alias function"
  crash=$1

  #
  # Search for indices that are using the alias we are trying to set
  #

  getAliasesCurl=`curl -XGET -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$cat/aliases" --write-out %{http_code} 2>/dev/null`
  http_code=`echo $catCurl | cut -d] -f2`
  if [[ "$http_code" != 200 ]]; then
    tryCrash $crash "Status code $http_code returned from catting aliases"
	echo "Status code $http_code returned from catting aliases"
  fi

  #
  # Extract index names that are using the alias from the above result
  #

  regex=""\""alias"\"":"\""$ALIAS_NAME"\"","\""index"\"":"\""([^"\""]+)"
  temp=`echo $getAliasesCurl|grep -Eo $regex | cut -d\" -f8`
  indexArr=(${temp// / })
  printIfTesting "Found ${#indexArr[@]} indices currently using alias $ALIAS_NAME: ${indexArr[@]}"

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
      tryCrash $crash "Failed to remove alias $ALIAS_NAME on index $index. Code: $http_code"
      echo "Failed to remove alias $ALIAS_NAME on index $index. Code: $http_code"
    fi
    printIfTesting "Removed alias $ALIAS_NAME on index $index"
  done
}

function createAlias {
  printIfTesting "Running create alias function"
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
    tryCrash $crash "Failed to create alias $ALIAS_NAME on index $INDEX_NAME. Code: $http_code"
    echo "Failed to create alias $ALIAS_NAME on index $INDEX_NAME. Code: $http_code"
  fi
  printIfTesting "Created alias $ALIAS_NAME on index $INDEX_NAME"
}

#
# Check to see if index already exists
#

printIfTesting "Checking to see if index $INDEX_NAME already exists..."
catCurl=`curl -X GET -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$cat/indices" --write-out %{http_code} 2>/dev/null`
http_code=`echo $catCurl | cut -d] -f2`
if [[ "$http_code" != 200 ]]; then
  failure "Status code $http_code returned while checking indices"
fi

if [[ $catCurl == *""\""index"\"":"\""$INDEX_NAME"\"""* ]]; then
  removeAliases false
  createAlias true
  success "Index already exists"
fi

#
# Create the index
#

printIfTesting "Creating index $INDEX_NAME with mappings..."
createIndexCurl=`curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "$IndexSettings" "$ES_IP$INDEX_NAME" --write-out %{http_code} 2>/dev/null`
http_code=`echo $catCurl | cut -d] -f2`
if [[ $createIndexCurl != '{"acknowledged":true}200' ]]; then
  failure "Failed to create index $INDEX_NAME. Code: $http_code"
fi

#
# If testing, create two indices that have the alias we are trying to set
#

if [ "$TESTING" = true ] ; then
    echo "Creating test indices..."
    peach=peach
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{}" "$ES_IP$peach" --write-out %{http_code}; echo " "
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [ { "\""add"\"" : { "\""index"\"" : "\""peach"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } } ]
    }" "$ES_IP$aliases" --write-out %{http_code}; echo " "
    pineapple=pineapple
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{}" "$ES_IP$pineapple" --write-out %{http_code}; echo " "
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [ { "\""add"\"" : { "\""index"\"" : "\""pineapple"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } } ]
    }" "$ES_IP$aliases" --write-out %{http_code}; echo " "
fi

removeAliases false

createAlias true

success "Index created successfully"
