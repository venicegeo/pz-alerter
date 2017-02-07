#setting GOPATH...
#export GOPATH=/home/vagrant/workspace/gostuff
#echo 'export GOPATH=/home/vagrant/workspace/gostuff' >>/home/vagrant/.bash_profile

#ENV variables for logger
#export VCAP_SERVICES='{"user-provided":[{"credentials":{"host":"192.168.44.44:9200","hostname":"192.168.44.44","port":"9200"},"label":"user-provided","name":"pz-elasticsearch","syslog_drain_url":"","tags":[]}]}'
#export PORT=14600
#export VCAP_APPLICATION='{"application_id": "14fca253-8081-402e-abf5-8fd40ddda81f","application_name": "pz-logger","application_uris": ["pz-logger.int.geointservices.io"],"application_version": "5f0ee99-q252c-4f8d-b241-bc3e22534afc","limits": {"disk": 1024,"fds": 16384,"mem": 512},"name": "pz-logger","space_id": "d65a0987-df00-4d69-a50b-657e52cb2f8e","space_name": "simulator-stage","uris": ["pz-logger.int.geointservices.io"],"users": null,"version": "5f0ee99d-252c-4f8d-b241-bc3e22534afc"}'

#start pz-logger app
/home/vagrant/workspace/gostuff/bin/pz-logger

#cd /home/vagrant/workspace/gostuff/bin
#./pz-logger