# Running pz-workflow locally

To run workflow, golang 1.7 and a valid go environment are required. Installation instructions can be found here: https://golang.org/doc/install

In order for workflow to successfully start it needs access to running ElasticSearch, Kafka, pz-servicecontroller, pz-idam services.
Additionally, the environment variable `LOGGER_INDEX` may be set; the value of this will be the name of the index in ElasticSearch for logger purposes. When running locally workflow will connect with ElasticSearch locally, however the `DOMAIN` environment variable must be set to the domain where the rest of Piazza is running in order to find pz-servicecontroller and pz-idam.

NOTE: pz-workflow cannot successfully execute triggers when running locally. There is currently no way of reaching the kafka service.

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
