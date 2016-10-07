#!/usr/bin/env bash
sudo apt-get update
sudo apt-get upgrade

# install openjdk-7 
#sudo apt-get purge openjdk*
#sudo apt-get -y install openjdk-7-jdk

alias list="ls -a"
 
#get golang
cd /usr/local
sudo wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz

#add go to path
echo 'export PATH=$PATH:/usr/local/go/bin' >>/home/vagrant/.bash_profile
export PATH=$PATH:/usr/local/go/bin


# install git
apt-get -y install git

#build and start the go app
mkdir /home/vagrant/workspace
cd /home/vagrant/workspace

#setting GOPATH...
export GOPATH=/home/vagrant/workspace/gostuff
#echo 'export GOPATH=/home/vagrant/workspace/gostuff' >>/home/vagrant/.bash_profile

#ENV variables for logger
export VCAP_SERVICES='{"user-provided":[{"credentials":{"host":"192.168.44.44:9200","hostname":"192.168.44.44","port":"9200"},"label":"user-provided","name":"pz-elasticsearch","syslog_drain_url":"","tags":[]}]}'
export PORT=14600
export VCAP_APPLICATION='{"application_id": "14fca253-8081-402e-abf5-8fd40ddda81f","application_name": "pz-logger","application_uris": ["pz-logger.int.geointservices.io"],"application_version": "5f0ee99-q252c-4f8d-b241-bc3e22534afc","limits": {"disk": 1024,"fds": 16384,"mem": 512},"name": "pz-logger","space_id": "d65a0987-df00-4d69-a50b-657e52cb2f8e","space_name": "simulator-stage","uris": ["pz-logger.int.geointservices.io"],"users": null,"version": "5f0ee99d-252c-4f8d-b241-bc3e22534afc"}'

#copying required set env script to profile.d for startup of the box
chmod 777 /vagrant/pzlogger/config/logger-env-variables.sh
cp /vagrant/pzlogger/config/logger-env-variables.sh /etc/profile.d/logger-env-variables.sh

#chmod 777 /vagrant/pzlogger/config/logger-startup.sh
#cp /vagrant/pzlogger/config/logger-startup.sh /etc/init.d/logger-startup.sh

#adding startup scripts to /etc/rc.local
cd /etc
echo '#!/bin/sh -e' > rc.local
echo 'su - root -c /home/vagrant/workspace/gostuff/bin/pz-logger &' >> rc.local
echo 'exit 0' >> rc.local

#echo getting pz-logger and trying to build it...
cd /home/vagrant/workspace
go get github.com/venicegeo/pz-logger
go install github.com/venicegeo/pz-logger

#start the app on initial box setup.
cd /home/vagrant/workspace/gostuff/bin/
nohup ./pz-logger &