#!/bin/sh


// zookeeper-server-start.sh ~/Downloads/kafka_2.10-0.8.2.0/config/zookeeper.properties
// kafka-server-start.sh ~/Downloads/kafka_2.10-0.8.2.0/config/server.properties
//
// kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic test
// kafka-topics.sh --list --zookeeper localhost:2181
//
// kafka-console-producer.sh --broker-list localhost:9092 --topic test
// kafka-console-consumer.sh --zookeeper localhost:2181 --topic test --from-beginning

zookeeper-server-start.sh ./config/zookeeper.properties &
kafka-server-start.sh ./config/server.properties &

