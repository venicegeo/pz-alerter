#/bin/sh

etId=$1

#query='{\"query\" : {\"bool\": {\"must\": [{\"match\" : {\"severity\" : 4}},{\"match\" : {\"problem\" : \"us-bbox\"}}]}}}'

query='{\"query\" :  {\"match\" : {\"problem\" : \"us-bbox\"}}}'

echo $query

cat > t <<foo
{
    "title": "my found-a-bad-telephone-number trigger",
    "condition": {
        "type": "$etId",
        "query": " $query ",
        "job": "do the thing!"
    }
}
foo

json=`cat t`

echo POST /triggers
echo "$json"

ret=`curl -S -s -XPOST -d "$json" http://pz-workflow.cf.piazzageo.io/v1/triggers`

echo RETURN:
echo $ret
