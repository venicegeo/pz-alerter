package main

import (
	//"bytes"
	//"errors"
	"fmt"
	//"github.com/mpgerlek/piazza-simulator/piazza"
	//"io/ioutil"
	"time"
	"testing"
)

var numRead int

const kafkaHost = "localhost:9092"
var kafka *Kafka

func doReads(t *testing.T, numReads *int) {
	c, err := kafka.NewConsumer()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	partitionConsumer, err := c.ConsumePartition("test3", 0, OffsetNewest)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	for  {
		msg := <-partitionConsumer.Messages()
		t.Logf("Consumed: offset:%d  value:%v", msg.Offset, string(msg.Value))
		*numReads++
	}

	t.Log("Reader done: %d", *numReads)
}


func doWrites(t *testing.T, id int, count int) {

	p, err := kafka.NewProducer()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// TODO: handle "err := <-w.Errors():"

	for n := 0; n<count; n++ {
		p.Input() <- NewMessage("test3", fmt.Sprintf("mssg %d from %d", n, id))
	}

	t.Logf("Writer done: %d", count)
}


func TestKafka(t *testing.T) {
	kafka = &Kafka{host: kafkaHost}

	var numReads1, numReads2 int

	go doReads(t, &numReads1)
	go doReads(t, &numReads2)

	n := 3
	go doWrites(t, 1, n)
	go doWrites(t, 2, n)

	time.Sleep(1 * time.Second)

	t.Log(numReads1,"---")
	t.Log(numReads2,"---")

	if numReads1 != n*2 {
		t.Fatalf("read1 count was %d, expected %d", numReads1, n*2)
	}
	if numReads2 != n*2 {
		t.Fatalf("read2 count was %d, expected %d", numReads2, n*2)
	}
}
