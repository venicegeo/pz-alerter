#/bin/sh


echo GET /alerts

ret=`curl -S -s -XGET -d "$json" http://pz-workflow.cf.piazzageo.io/v1/alerts`

echo RETURN:
echo $ret
