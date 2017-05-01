# Running pz-workflow locally

To run workflow golang 1.7 and a valid go environment are required.

Create the directories 'src/venicegeo' in $GOPATH

From '$GOPATH/src/venicegeo/' clone pz-workflow with 'git clone https://github.com/venicegeo/pz-workflow'

To simply build, from '$GOPATH/src/venicegeo/pz-workflow/' execute 'go build'

To run execute './pz-workflow'

To run pz-workflow without building execute 'go run main.go`

In order for workflow to successfully start, it needs access to running ElasticSearch, Kafka, pz-servicecontroller, pz-idam services.
