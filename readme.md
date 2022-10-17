![header](simp_mq.png)

# SimpMQ
##### a dead simple message queue with no dependencies and doesn't support any storage (not yet), or follow any protocol, written in go.

## _there are two compenents_
### SimpBroker
broker accepts subscriptions and distributes published messages to subscribers

#### how to use broker
to run a broker named "demo_broker" on port "8080", expects clients to authenticate with broker
```go
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
```

to close the broker and exist
```go
broker.close()
```
### SimpClient
simp client makes it easier for mq clients to subscribe and publish to topics, as well as recieve published messages
#### how to use client
start a client named "sub_client" and connect it to broker running on "localhost:8080" which requires a authentication token "password"
```go
client := simpmq.SimpClient{
	Id:             "sub_client",
	SimpBrokerHost: "localhost:8080",
	Token:          "password",
}
```
now actually connect to the server
```go
err := client.ConnectToServer()
if err != nil {
	return err
}
```
subscribe to a topic called "demo_topic"
```go
subscriber := make(chan []byte)
err = client.Subscribe("demo_topic", func(bytes []byte) {
	subscriber <- bytes
})
if err != nil {
	return err
}
```
publish to a topic called "demo_topic"
```go
err := client.Publish("demo_topic", []byte("message to send"))

if err != nil {
	//fail...
}
```
