#!/usr/bin/env bash
sudo apt-get update
sudo apt-get upgrade

# install openjdk-7 
sudo apt-get purge openjdk*
sudo apt-get -y install openjdk-7-jdk

#alias list="ls -a"
 
#get golang
cd /usr/local
sudo wget https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.6.2.linux-amd64.tar.gz

#add go to path
#cd /etc/profile.d
#sudo echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile.d/gosetup.sh
echo 'export PATH=$PATH:/usr/local/go/bin' >>/home/vagrant/.bash_profile
export PATH=$PATH:/usr/local/go/bin
#echo $PATH

echo go help.
go help

#apt-get install git-all
#apt-get install git htop -y -q
apt-get -y install git

#su - vagrant

#cd /etc/profile.d
#touch addgolangtopath.sh
#echo 'export GOPATH=/home/vagrant/workspace/gostuff' >> addgolangtopath.sh

#apt-get install golang-go
#apt-get -y install golang-go

#build and start the go app
#echo Creating new workspace directory at home directory
mkdir /home/vagrant/workspace
cd /home/vagrant/workspace

#echo Cloning pz-logger repository...
#git clone https://github.com/venicegeo/pz-logger.git
#cd pz-logger
#echo Creating gogo directory within pz-logger for GOPATH
#mkdir gogo

#echo setting GOPATH...
#export GOPATH=$(pwd -P)/workspace/pz-logger/gogo
export GOPATH=/home/vagrant/workspace/gostuff
#echo 'export GOPATH=/home/vagrant/workspace/gostuff' >>/home/vagrant/.bash_profile

#echo 'export GOPATH=$(pwd -P)/workspace/pz-logger/gogo' >>/home/vagrant/.bash_profile
#echo Installing pz-logger go install...

#echo go help..........................................
#go help

#echo getting pz-uuidgen and trying to build it.....................
go get github.com/venicegeo/pz-workflow
go install github.com/venicegeo/pz-workflow/server

#start the app
cd /home/vagrant/workspace/gostuff/bin
echo List of workflow executables:
ls -la
#./pz-logger
