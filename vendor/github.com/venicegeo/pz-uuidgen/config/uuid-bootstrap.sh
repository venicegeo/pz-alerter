#!/usr/bin/env bash
sudo apt-get update
sudo apt-get upgrade
 
# install golang
cd /usr/local
sudo wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz

# add golang to path
echo 'export PATH=$PATH:/usr/local/go/bin' >>/home/vagrant/.bash_profile
export PATH=$PATH:/usr/local/go/bin

# install git
apt-get -y install git

# creating new workspace directory at home directory
mkdir /home/vagrant/workspace
cd /home/vagrant/workspace

# setting env variables
export GOPATH=/home/vagrant/workspace/gostuff
export VCAP_SERVICES='{"user-provided": [{"credentials": {"host": "192.168.44.44:9200","hostname": "192.168.44.44","port": "9200"},"label": "user-provided","name": "pz-elasticsearch","syslog_drain_url": "","tags": []},{"credentials": {"host": "192.168.46.46:14600","hostname": "192.168.46.46","port": "14600"},"label": "user-provided","name": "pz-logger","syslog_drain_url": "","tags": []}]}'
export VCAP_APPLICATION='{"application_id": "fe5dfc8d-e36e-4f21-9223-2ed4f7a984dd","application_name": "pz-uuidgen","application_uris": ["pz-uuidgen.int.geointservices.io","pz-uuidgen-Sprint03-74-g7862a67.int.geointservices.io"],"application_version": "f3905ce7-52f3-4d35-8309-1003963250ca","limits": {"disk": 1024,"fds": 16384,"mem": 512},"name": "pz-uuidgen","space_id": "5f97f401-4277-4a13-bbd9-5e5ff62f21a2","space_name": "int","uris": ["pz-uuidgen.int.geointservices.io","pz-uuidgen-Sprint03-74-g7862a67.int.geointservices.io"],"users": null,"version": "f3905ce7-52f3-4d35-8309-1003963250ca"}'
export PORT=14800

#copying required set env script to profile.d for startup of the box
chmod 777 /vagrant/uuid/config/uuid-env-variables.sh
cp /vagrant/uuid/config/uuid-env-variables.sh /etc/profile.d/uuid-env-variables.sh

cd /etc
echo '#!/bin/sh -e' > rc.local
echo 'su - root -c /home/vagrant/workspace/gostuff/bin/pz-uuidgen &' >> rc.local
echo 'exit 0' >> rc.local

# getting pz-uuidgen and trying to build it
go get github.com/venicegeo/pz-uuidgen
go install github.com/venicegeo/pz-uuidgen

# start the app
cd /home/vagrant/workspace/gostuff/bin
echo List of built executables:
ls -la

#start the app on initial box setup.
cd /home/vagrant/workspace/gostuff/bin/
nohup ./pz-uuidgen &