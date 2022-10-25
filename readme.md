![header](simp_mq.png)

# SimpMQ
#### a dead simple message queue with no dependencies and doesn't support any storage (not yet), or follow any protocol, written in go.
#

>```sh
>WARNING:
>
>Even though it is working, as it is in early developement, expect breaking changes
>```

## Installation for use in go code

To install simp_mq package, you need to install Go and set your Go workspace first.

1. You first need [Go](https://golang.org/) installed (**version 1.16+ is required**)

command to install and run a local SimpBroker
```sh
go install github.com/ondbyte/simp_mq
```
after which you can use `simp_mq` command, which has only one sub command `startbroker`, which is used to, you guessed it to start a SimpBroker instance to which you can connect to from any SimpClient,
example command to start a SimpBroker
```sh
simp_mq startbroker --name xyz --port 8080 --token password
```
other options available
```sh
   --name value, -n value        name of the SimpBroker instance (default: "demo_simp_broker")
   --port value, -p value        set port for SimpBroker to run on (default: 8081)
   --token value, -t value       SimpBroker will authenticate a client using this token (default: "password")
   --buffersize value, -b value  maximum size of the each message exchanged between client and broker (default: 2048)
   --authwait value, -w value    SimpBroker will wait these milliseconds for authentication from a new connection, after which connection will be dropped (default: 10s)
```

command to install simp_broker.
```sh
go get -u github.com/ondbyte/simp_mq/simp_broker
```

command to install simp_client.
```sh
go get -u github.com/ondbyte/simp_mq/simp_client
```

2. Import it in your code:

```go
import "github.com/ondbyte/simp_mq/simp_broker"
//or
import "github.com/ondbyte/simp_mq/simp_client"
```
## How to use

### SimpBroker
broker accepts subscriptions and distributes published messages to subscribers

#### how to use broker
to run a broker named "demo_broker" on port "8080", expects clients to authenticate with broker
```go
broker := &simp_broker.SimpBroker{
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
client := &simp_client.SimpClient{
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
