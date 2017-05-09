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

if [ "$ALIAS_NAME" = "" ]; then
  failure "Please specify an alias name as argument 1"
fi

if [ "$ES_IP" = "" ]; then
  failure "Please specify the elasticsearch ip as argument 2"
fi 

if [[ $ES_IP != *"/" ]]; then
  ES_IP="$ES_IP/"
fi

if [ "$TESTING" = "" ]; then
  TESTING=false
fi

function tryCrash {
	if [ "$1" = true ] ; then
	  failure "$2"
	fi
}

function handleAliases {
  printIfTesting "Running handle alias function"
  crash=true

  #
  # Search for indices that are using the alias we are trying to set
  #

  getAliasesCurl=`curl -XGET -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$cat/aliases?format=json" --write-out %{http_code} 2>/dev/null`
  http_code=${getAliasesCurl: -3}
  if [ "$http_code" -ne 200 ]; then
    tryCrash $crash "Status code $http_code returned from catting aliases"
	printIfTesting "Status code $http_code returned from catting aliases"
  fi

  #
  # Extract index names that are using the alias from the above result
  #

  regex=""\""alias"\"":"\""$ALIAS_NAME"\"","\""index"\"":"\""([^"\""]+)"
  temp=`echo $getAliasesCurl|grep -Eo $regex | cut -d\" -f8`
  indexArr=(${temp// / })
  if [ "$TESTING" = true ] ; then
    echo "Found ${#indexArr[@]} indices currently using alias $ALIAS_NAME: ${indexArr[@]}"
  fi
  if [ "${#indexArr[@]}" -eq 1 ] ; then
    if [ "${indexArr[0]}" = $INDEX_NAME ]; then
	  printIfTesting "Alias already exists on index"
	  return
    fi
  fi
  declare -a actions=()
  actionsLength=0
  for index in ${indexArr[@]}
  do
    actions[actionsLength]="{"\""remove"\"":{"\""index"\"":"\""$index"\"","\""alias"\"":"\""$ALIAS_NAME"\""}}"
    let actionsLength+=1
  done
  actions[actionsLength]="{"\""add"\"":{"\""index"\"":"\""$INDEX_NAME"\"","\""alias"\"":"\""$ALIAS_NAME"\""}}"
  let actionsLength+=1
  concatCount=0
  total=""
  for action in ${actions[@]}
  do
    total=$total$action
    let aLt=$actionsLength-1
	if [ $concatCount -lt $aLt ]; then
	  total="$total,"
	fi
	let concatCount+=1
  done
  total="{"\""actions"\"":[$total]}"
  aliasCurl=`curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "$total" "$ES_IP$aliases" --write-out %{http_code} 2>/dev/null`
  http_code=${aliasCurl: -3}
    if [ "$aliasCurl" != '{"acknowledged":true}200' ]; then
      tryCrash $crash "Failed to set up aliases"
    else
      printIfTesting "Successfully set up aliases"
	fi
}

#
# Check to see if index already exists
#

printIfTesting "Checking to see if index $INDEX_NAME already exists..."
catCurl=`curl -X GET -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$cat/indices?format=json" --write-out %{http_code} 2>/dev/null`
http_code=${catCurl: -3}
if [ "$http_code" -ne 200 ]; then
  failure "Status code $http_code returned while checking indices"
fi

if [[ $catCurl == *""\""index"\"":"\""$INDEX_NAME"\"""* ]]; then
  handleAliases
  success "Index already exists"
fi

#
# Create the index
#

printIfTesting "Creating index $INDEX_NAME with mappings..."
createIndexCurl=`curl -X PUT -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "$IndexSettings" "$ES_IP$INDEX_NAME" --write-out %{http_code} 2>/dev/null`
http_code=${createIndexCurl: -3}
if [ $createIndexCurl != '{"acknowledged":true,"shards_acknowledged":true}200' ]; then
  failure "Failed to create index $INDEX_NAME. Code: $http_code"
fi

#
# If testing, create two indices that have the alias we are trying to set
#

if [ "$TESTING" = true ] ; then
    echo "Creating test indices..."
    peach=peach
    curl -X PUT -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{}" "$ES_IP$peach" --write-out %{http_code}; echo " "
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [ { "\""add"\"" : { "\""index"\"" : "\""peach"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } } ]
    }" "$ES_IP$aliases" --write-out %{http_code}; echo " "
    pineapple=pineapple
    curl -X PUT -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{}" "$ES_IP$pineapple" --write-out %{http_code}; echo " "
    curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d "{
    "\""actions"\"" : [ { "\""add"\"" : { "\""index"\"" : "\""pineapple"\"", "\""alias"\"" : "\""$ALIAS_NAME"\"" } } ]
    }" "$ES_IP$aliases" --write-out %{http_code}; echo " "
fi

handleAliases

if [ "$TESTING" = true ] ; then	
	echo "Deleting test indices..."
	peach=peach
    curl -X DELETE -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$peach" --write-out %{http_code}; echo " "
   	pineapple=pineapple
    curl -X DELETE -H "Content-Type: application/json" -H "Cache-Control: no-cache" "$ES_IP$pineapple" --write-out %{http_code}; echo " "
fi

success "Index created successfully"
