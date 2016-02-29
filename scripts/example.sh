#/bnin/sh

addr="http://pz-workflow.cf.piazzageo.io"

# health check
curl -X GET  $addr

echo ; echo

# make a trigger
curl -X POST -d '{"title":"Test", "condition": {"type":"Foo","query":"the query"}, "job":{"task":"do something"}}' $addr/v1/triggers

echo ; echo

# show all the things
curl -X GET $addr/v1/triggers

# issue an event
curl -X POST -d '{"type":"Foo","date":"2016-02-16T21:20:48.052Z","data":{}}' $addr/v1/events

echo ; echo

# show all the events
curl -X GET $addr/v1/events

#
curl -X GET $addr/v1/alerts
