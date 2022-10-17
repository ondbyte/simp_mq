package simpmq_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	simpmq "yadunandan.xyz/simp_mq"
)

var (
	toSend = []string{
		"yadu", "is", "awesome", "stop",
	}
)

func TestMQ(t *testing.T) {
	ch := make(chan bool)
	broker, err := startSimpBroker(t)
	if err != nil {
		t.Error(err)
		broker.Close()
		close(ch)
		return
	}
	time.Sleep(time.Second)
	go func() {
		err := startSubscriptionClient(t)
		if err != nil {
			panic(err)
		}
		broker.Close()
		close(ch)
	}()
	time.Sleep(time.Second)
	go func() {
		err := startPublishClient(t)
		if err != nil {
			panic(err)
		}
	}()
	<-ch
}

func startSimpBroker(t *testing.T) (*simpmq.SimpBroker, error) {
	broker := &simpmq.SimpBroker{
		Id:   "demo_broker",
		Port: "8080",
		Authenticator: func(deets *simpmq.AuthDetails) error {
			if deets.Token == "password" {
				return nil
			}
			return errors.New("failed to authenticate")
		},
	}

	err := broker.Serve()
	if err != nil {
		return nil, err
	}
	return broker, nil
}
func startSubscriptionClient(t *testing.T) error {
	client := simpmq.SimpClient{
		Id:             "sub_client",
		SimpBrokerHost: "localhost:8080",
		Token:          "password",
	}

	err := client.ConnectToServer()
	if err != nil {
		return err
	}
	subscriber := make(chan []byte)
	err = client.Subscribe("demo_topic", func(bytes []byte) {
		subscriber <- bytes
	})
	if err != nil {
		return err
	}
	recd := make([][]byte, 0)
	for {
		message := <-subscriber
		fmt.Println(string(message))
		recd = append(recd, message)
		if string(message) == "stop" {
			break
		}
	}

	err = client.UnSubscribe("demo_topic")
	close(subscriber)
	client.Close()

	if err != nil {
		t.Errorf("failed to unsubscribe")
	}

	if len(recd) != len(toSend) {
		t.Errorf("failed to recieve sent messages on client, number of recieved msg is %d but should be %d", len(recd), len(toSend))
	}
	client.Close()
	return nil
}

func startPublishClient(t *testing.T) error {
	client := simpmq.SimpClient{
		Id:             "pub_client",
		SimpBrokerHost: "localhost:8080",
		Token:          "password",
	}

	err := client.ConnectToServer()
	if err != nil {
		return err
	}

	recd := make([]error, 0)
	for _, message := range toSend {
		err := client.Publish("demo_topic", []byte(message))

		if err != nil {
			recd = append(recd, err)
		}
	}

	if len(recd) > 0 {
		t.Errorf("failed to send message from publisher")
	}
	client.Close()
	return nil
}

func TestError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(fmt.Sprintf("%+v", r))
		}
	}()
	panic("what")
}

func TestTime(t *testing.T) {
	fmt.Println(time.Now())
	fmt.Println(time.Now())
	fmt.Println(time.Now())
}

func TestTemp(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8}

	wg := sync.WaitGroup{}

	var result []int

	for i, num := range input {
		wg.Add(1)
		go func(num, i int) {
			if num%2 == 0 {
				result = append(result, num)
			}

			wg.Done()
		}(num, i)
	}
	wg.Wait()

	fmt.Println(result)
}
