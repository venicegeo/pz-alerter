# Running pz-workflow locally

To run workflow golang 1.7 and a valid go environment are required.

In order for workflow to successfully start, it needs access to running ElasticSearch, Kafka, pz-servicecontroller, pz-idam services.
Additionally, the environment variable `LOGGER_INDEX` may be set; the value of this will be the name of the index in ElasticSearch for logger purposes.

Execute:
```
mkdir $GOPATH/src
mkdir $GOPATH/src/github.com/
mkdir $GOPATH/src/github.com/venicegeo
cd $GOPATH/src/github.com/venicegeo/
git clone https://github.com/venicegeo/pz-workflow
go build
./pz-workflow
```
