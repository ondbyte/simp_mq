package simpmq

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"yadunandan.xyz/simp_mq/slog"
)

type SimpClient struct {
	Id                 string                          //id
	SimpBrokerHost     string                          //host address of the broker,mostly a local host
	Token              string                          //token used to authenticate with the broker
	subscriptions      map[string]SubscribtionListener //all subscriber according to topic
	waitingForSubUnSub map[string]chan bool
	waitingForPubAck   map[string]chan bool
	conn               *SimpServerConn //connection to the server
	connectedToServer  chan bool       //usd to close all dependencies
	ConnectedToServer  bool            //whether connection is active
}

//non blocking
//establishes a connection to broker
//call Close to disconnect
func (client *SimpClient) ConnectToServer() (err error) {
	client.waitingForSubUnSub = make(map[string]chan bool)
	client.subscriptions = make(map[string]SubscribtionListener)
	client.waitingForPubAck = make(map[string]chan bool)
	conn, err := net.Dial("tcp", client.SimpBrokerHost)
	if err != nil {
		return err
	}
	simpConn := &SimpServerConn{NetConn: conn, BufferSize: 1024, AuthDetails: &AuthDetails{
		Token:    client.Token,
		ClientID: client.Id,
	}}

	err = simpConn.authenticateWithBroker()

	if err != nil {
		return err
	}
	fmt.Printf("SimpClient with id %s is active\n", client.Id)
	client.conn = simpConn
	// start a go routine loop to wait for next data from server
	go func() {
		for {
			data, err := simpConn.nextDataFromConnection()
			if err == nil {
				switch data.Type {

				case pub:
					{
						//handle a published message
						deets, err := data.GetPubDetails()
						if err == nil {
							listener, waiting := client.subscriptions[deets.Topic]
							if waiting {
								listener(deets.Data)
							}
						} else {
							slog.Warn("unable to get pub details code: xyz122")
						}
					}

				case subAck, unsubAck:
					{
						//handle a subscribe acknowledgement message
						ch, waiting := client.waitingForSubUnSub[data.ID]
						if waiting {
							ch <- true
						}
					}
				case pubAck:
					{
						//handle a publish acknowledgement message
						ch, waiting := client.waitingForPubAck[data.ID]
						if waiting {
							ch <- true
						}
					}
				}
			}
		}
	}()
	//wait for closing event
	go func() {
		_, more := <-client.connectedToServer
		if !more {
			client.ConnectedToServer = false
			conn.Close()
			fmt.Printf("[%s] disconnected from broker %s and exited\n", client.Id, conn.RemoteAddr())
		}
	}()
	client.ConnectedToServer = true
	client.connectedToServer = make(chan bool)
	return nil
}

func (client *SimpClient) Close() {
	if client.ConnectedToServer {
		close(client.connectedToServer)
	}
}

type SubscribtionListener func([]byte)

//subcribe to the given topic, messages will be delivered on the listener
//completes when a subscription acknowledgement is recieved, which is not guaranteed in real life conditions
func (client *SimpClient) Subscribe(topic string, listener SubscribtionListener) error {
	_, alreadyTrying := client.waitingForSubUnSub[topic]
	if alreadyTrying {
		return fmt.Errorf("subscription/unsubscription request aleady sent for topic %s, waiting for acknowledgement from broker", topic)
	}
	_, alreadySubscribed := client.subscriptions[topic]
	if alreadySubscribed {
		return fmt.Errorf("already subscribed to topic %s, waiting for new messages to arrive", topic)
	}
	id := string(rune(time.Now().UnixNano()))
	deets := &SubDetails{Topic: topic}
	payload, err := deets.Marshal()
	if err != nil {
		return err
	}
	err = client.conn.respond(&SimpData{Type: sub, ID: id, Payload: payload})
	if err != nil {
		return err
	}
	ch := make(chan bool)
	client.waitingForSubUnSub[id] = ch
	_, subscribed := <-ch
	delete(client.waitingForSubUnSub, id)

	if !subscribed {
		return fmt.Errorf("failed to subscribe to %s", topic)
	}

	client.subscriptions[topic] = listener
	close(ch)
	return nil
}

//subcribe to the given topic, messages will be delivered on the listener
func (client *SimpClient) UnSubscribe(topic string) error {
	_, alreadyTrying := client.waitingForSubUnSub[topic]
	if alreadyTrying {
		return fmt.Errorf("subscription/unsubscription request aleady sent for topic %s, waiting for acknowledgement from broker", topic)
	}
	_, alreadySubscribed := client.subscriptions[topic]
	if !alreadySubscribed {
		return fmt.Errorf("not subscribed to topic %s to unsubscribe", topic)
	}
	id := string(rune(time.Now().UnixNano()))
	payload, err := json.Marshal(&SubDetails{Topic: topic})
	if err != nil {
		return err
	}
	err = client.conn.respond(&SimpData{Type: unsub, ID: id, Payload: payload})
	if err != nil {
		return err
	}
	ch := make(chan bool)
	client.waitingForSubUnSub[id] = ch
	_, unsubscribed := <-ch
	delete(client.waitingForSubUnSub, id)
	if !unsubscribed {
		return fmt.Errorf("failed to unsubscribe to topic %s", topic)
	}
	close(ch)
	return nil
}

//get a channel to publish to a topic, add a message to the returned pb channel to publish a message
//closing the channel will stop you from publishing to this topic which also closes the ackknowledgment channel
//you'll recieve ID of the message acknowledgemnt by broker on ack channel
func (client *SimpClient) Publish(topic string, payload []byte) error {
	id := topic + string(rune(time.Now().UnixNano()))
	payload, err := json.Marshal(&PubDetails{Topic: topic, Data: payload})
	if err != nil {
		return err
	}
	err = client.conn.respond(&SimpData{Type: pub, ID: id, Payload: payload})
	if err != nil {
		return err
	}

	ch := make(chan bool)
	client.waitingForPubAck[id] = ch
	_, open := <-ch
	close(ch)
	delete(client.waitingForPubAck, id)
	if !open {
		return fmt.Errorf("failed to recieve acknowledgement for message")
	}
	return nil
}
