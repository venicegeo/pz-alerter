#!/usr/bin/env bash
sudo apt-get update
sudo apt-get upgrade

# install openjdk-7 
sudo apt-get purge openjdk*
sudo apt-get -y install openjdk-7-jdk

#get golang
cd /usr/local
sudo wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz

#add go to path
echo 'export PATH=$PATH:/usr/local/go/bin' >>/home/vagrant/.bash_profile
export PATH=$PATH:/usr/local/go/bin

echo go help.
go help

apt-get -y install git

#build and start the go app
#echo Creating new workspace directory at home directory
mkdir /home/vagrant/workspace
cd /home/vagrant/workspace

#echo setting GOPATH...
export GOPATH=/home/vagrant/workspace/gostuff
#echo 'export GOPATH=/home/vagrant/workspace/gostuff' >>/home/vagrant/.bash_profile

#echo getting pz-uuidgen and trying to build it.....................
go get github.com/venicegeo/pz-workflow
go install github.com/venicegeo/pz-workflow

#start the app
cd /home/vagrant/workspace/gostuff/bin

echo List of workflow executables:
ls -la

./pz-workflow
