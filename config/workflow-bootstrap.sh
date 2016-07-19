#!/usr/bin/env bash
sudo apt-get update
sudo apt-get upgrade

# get golang
cd /usr/local
sudo wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz

# add go to path
echo 'export PATH=$PATH:/usr/local/go/bin' >>/home/vagrant/.bash_profile
export PATH=$PATH:/usr/local/go/bin

apt-get -y install git

# build and start the go app
mkdir /home/vagrant/workspace
cd /home/vagrant/workspace

# setting some env variables
export GOPATH=/home/vagrant/workspace/gostuff
export VCAP_SERVICES='{"user-provided": [{"credentials": {"host": "192.168.44.44:9200","hostname": "192.168.44.44","port": "9200"},"label": "user-provided","name": "pz-elasticsearch","syslog_drain_url": "","tags": []},{"credentials": {"host": "192.168.46.46:14600","hostname": "192.168.46.46","port": "14600"},"label": "user-provided","name": "pz-logger","syslog_drain_url": "","tags": []},{"credentials": {"host": "192.168.48.48:14800","hostname": "192.168.48.48","port": "14800"},"label": "user-provided","name": "pz-uuidgen","syslog_drain_url": "","tags": []}]}'
export PORT=14400
export VCAP_APPLICATION='{"application_id": "14fca253-8087-402e-abf5-8fd40ddda81f","application_name": "pz-workflow","application_uris": ["pz-workflow.int.geointservices.io"],"application_version": "5f0ee99d-252c-4f8d-b241-bc3e22534afc","limits": {"disk": 1024,"fds": 16384,"mem": 512},"name": "pz-workflow","space_id": "d65a0987-df00-4d69-a50b-657e52cb2f8e","space_name": "simulator-stage","uris": ["pz-workflow.int.geointservices.io"],"users": null,"version": "5f0ee99d-252c-4f8d-b241-bc3e22534afc"}'

# getting pz-workflow and trying to build it
go get github.com/venicegeo/pz-workflow
go install github.com/venicegeo/pz-workflow

#copying required set env script to profile.d for startup of the box
chmod 777 /vagrant/workflow/config/workflow-env-variables.sh
cp /vagrant/workflow/config/workflow-env-variables.sh /etc/profile.d/workflow-env-variables.sh

#adding startup scripts to /etc/rc.local
cd /etc
echo '#!/bin/sh -e' > rc.local
echo 'su - root -c /home/vagrant/workspace/gostuff/bin/pz-workflow &' >> rc.local
echo 'exit 0' >> rc.local

#start the app
cd /home/vagrant/workspace/gostuff/bin
nohup ./pz-workflow &
