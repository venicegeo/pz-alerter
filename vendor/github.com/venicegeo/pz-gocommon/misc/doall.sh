#!/bin/sh

set -e
set -x

go test -v github.com/venicegeo/pz-gocommon/gocommon

go test -v github.com/venicegeo/pz-gocommon/elasticsearch
cd ~/venicegeo/pz-gocommon/elasticsearch/systest
go test -v

go test -v github.com/venicegeo/pz-gocommon/kafka
#cd ~/venicegeo/pz-gocommon/kafka/systest
#go test -v

go test -v github.com/venicegeo/pz-logger/logger
cd ~/venicegeo/pz-logger/systest
go test -v sys_test.go
go build github.com/venicegeo/pz-logger

go test -v github.com/venicegeo/pz-uuidgen/uuidgen
cd ~/venicegeo/pz-uuidgen/systest
go test -v sys_test.go
go build github.com/venicegeo/pz-uuidgen

go test -v github.com/venicegeo/pz-workflow/workflow
cd ~/venicegeo/pz-workflow/systest
go test -v sys_test.go
sh demo.sh
go build github.com/venicegeo/pz-workflow

