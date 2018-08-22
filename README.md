# pz-workflow
The Piazza workflow service enables the construction and use of "event" notifications to enable simple "if-this-happens-then-do-that" workflows. This is done though an HTTP API.

## Requirements
Before building and running the pz-workflow project, please ensure that the following components or piazza services are available and/or installed, as necessary:
* [Go](https://golang.org/doc/install) v1.7 or later
* [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) (for checking out repository source)
- [Glide](https://glide.sh)
* [ElasticSearch](https://www.elastic.co/)
* [Apache Kafka](https://kafka.apache.org/quickstart)

## Setup, Configuring & Running

### Setup
Create the directory the repository must live in, and clone the git repository:

    $ mkdir -p $GOPATH/src/github.com/venicegeo
    $ cd $GOPATH/src/github.com/venicegeo
    $ git clone git@github.com:venicegeo/pz-workflow.git
    $ cd pz-workflow

Set up Go environment variables

To function right, Go must have some environment variables set. Run the `go env`
command to list all relevant environment variables. The two most important
variables to look for are `GOROOT` and `GOPATH`.

- `GOROOT` must point to the base directory at which Go is installed
- `GOPATH` must point to a directory that is to serve as your development
  environment. This is where this code and dependencies will live.

To quickly verify these variables are set, run the command from terminal:

	$ go env | egrep "GOPATH|GOROOT"

### Configuring
As noted in the Requirements section, the pz-workflow project needs access to a running, local ElasticSearch instance.
Additionally, the environment variable `LOGGER_INDEX` must be set; the value of this will be the name of the index in ElasticSearch containing logs. When running locally, workflow will connect with ElasticSearch locally, however the `DOMAIN` environment variable must be set to the domain where the rest of Piazza is running in order to find [pz-servicecontroller](https://github.com/venicegeo/pz-servicecontroller) and [pz-idam](https://github.com/venicegeo/pz-idam).

> __Note:__ pz-workflow cannot successfully execute triggers when running locally. There is currently no way of reaching the kafka service.

## Installing, Building, Running & Unit Tests

### Install dependencies

This project manages dependencies by populating a `vendor/` directory using the
glide tool. If the tool is already installed, in the code repository, run:

    $ glide install -v

This will retrieve all the relevant dependencies at their appropriate versions
and place them in `vendor/`, which enables Go to use those versions in building
rather than the default (which is the newest revision in Github).

> **Adding new dependencies.** When adding new dependencies, simply installing
  them with `go get <package>` will fetch their latest version and place it in
  `$GOPATH/src`. This is undesirable, since it is not repeatable for others.
  Instead, to add a dependency, use `glide get <package>`, which will place it
  in `vendor/` and update `glide.yaml` and `glide.lock` to remember its version.

### Build the project
To build `pz-workflow`, run `go install` from the project directory. To build it from elsewhere, run:

	$ go install github.com/venicegeo/pz-workflow

This will build and place a statically-linked Go executable at `$GOPATH/bin/pz-workflow`.

### Running locally

	$ go build
	$ ./pz-workflow

