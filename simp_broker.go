package simpmq

import (
	"fmt"
	"net"
	"time"
)

type SubScribers struct {
	all map[string]map[string]*SimpClientConn
}

func (SubScribers *SubScribers) init() {
	SubScribers.all = make(map[string]map[string]*SimpClientConn)
}

func (SubScribers *SubScribers) addForTopic(topic string, simpConn *SimpClientConn) {
	all := SubScribers.all[topic]
	if all == nil {
		all = make(map[string]*SimpClientConn)
	}
	all[simpConn.Id] = simpConn
	SubScribers.all[topic] = all
}

func (SubScribers *SubScribers) removeForTopic(topic string, simpConn *SimpClientConn) {
	all := SubScribers.all[topic]
	if all == nil {
		return
	}
	delete(all, simpConn.Id)
	SubScribers.all[topic] = all
}

//a simple broker which you can publish to subscribe to
type SimpBroker struct {
	//unique id
	Id string
	//port the broker must try to run on or fail
	Port string
	//topic subscribers
	subscribers *SubScribers
	//connectsions with their ids
	allConnections map[string]*SimpClientConn
	//internal channel recieves event when the server gets closed so any go routines depended on the server can close
	serverClosingEvent chan bool
	//whether the broker is running
	Running bool

	//max size of the message
	MaxMessageBuffer uint
	//validate token from a client for a successful connection
	Authenticator Authenticator
	//if no authentication data is recieved from a client, connection will be dropped after this duration
	DropNoAuthConnectionAfter time.Duration
}

//non blocking,
//starts a SimpBroker, ready for accepting new connections,
//Use SimpClient to access the broker, returns an error if broker fails to serve.
//call close when you wrap up.
func (broker *SimpBroker) Serve() (err error) {
	if broker.MaxMessageBuffer == 0 {
		broker.MaxMessageBuffer = 1024
	}
	broker.subscribers = &SubScribers{}
	broker.subscribers.init()
	broker.allConnections = make(map[string]*SimpClientConn)
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", broker.Port))
	if err != nil {
		return err
	}

	go func() {
		for {
			//wait for new connection
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
			}
			//new tcp connection
			broker.newIncomingConnection(conn)

		}
	}()
	go func() {
		fmt.Printf("SimpBroker is running on port %s\n", ln.Addr().String())
		for {
			_, more := <-broker.serverClosingEvent
			if !more {
				broker.Running = false
				ln.Close()
				fmt.Println("SimpMQ has shut down")
			}
		}
	}()
	broker.Running = true
	broker.serverClosingEvent = make(chan bool)
	return nil
}

//handles any new connections from clients
func (broker *SimpBroker) newIncomingConnection(conn net.Conn) {
	go func() {
		simpConn := &SimpClientConn{NetConn: conn, BufferSize: broker.MaxMessageBuffer, Authenticator: broker.Authenticator}
		broker.allConnections[simpConn.Id] = simpConn
		err := broker.authenticateNewSimpConnection(simpConn)
		if err != nil {
			fmt.Println(err)
			return
		}
		broker.afterAuthLoopForConn(simpConn)
	}()
}

//authenticateNewSimpConnection attempts to authenticate the simpConn, max waiting time for a client to send
//authentication information can be provided to the SimpBroker instance after which the connection will be failed
func (broker *SimpBroker) authenticateNewSimpConnection(simpConn *SimpClientConn) (err error) {
	authData, err := simpConn.authenticateWithClient()

	if err != nil {
		return err
	} else {
		authData.Type = authAck
		err = simpConn.respond(authData)
		if err != nil {
			simpConn.close()
			return err
		}
	}
	return nil
}

//handles further data after authentication of the connection
func (broker *SimpBroker) afterAuthLoopForConn(simpConn *SimpClientConn) (err error) {
	for {
		nextData, err := simpConn.nextDataFromConnection()
		if err == nil {
			switch nextData.Type {
			case pub:
				{
					deets, err := nextData.GetPubDetails()
					if err != nil {
						fmt.Println("theres error getting pub details simp_broker:afterAuthLoopForConn()")
					}
					for _, subscriber := range broker.subscribers.all[deets.Topic] {
						subscriber.respond(nextData)
					}
					//send acknkowledge
					nextData.Type = pubAck
					err = simpConn.respond(nextData)
					if err != nil {
						fmt.Println("error responding")
					}
					break
				}
			case sub:
				{
					deets, err := nextData.GetSubDetails()
					if err != nil {
						fmt.Println("theres error getting sub details simp_broker:afterAuthLoopForConn()")
					}
					broker.subscribers.addForTopic(deets.Topic, simpConn)
					//send acknkowledge
					nextData.Type = subAck
					err = simpConn.respond(nextData)
					if err != nil {
						fmt.Println("error responding")
					}
					break
				}
			case unsub:
				{
					deets, err := nextData.GetSubDetails()
					if err != nil {
						fmt.Println("theres error getting sub details simp_broker:afterAuthLoopForConn()")
					}
					broker.subscribers.removeForTopic(deets.Topic, simpConn)
					//send acknkowledge
					nextData.Type = unsubAck
					err = simpConn.respond(nextData)
					if err != nil {
						fmt.Println("error responding")
					}
					break
				}
			case auth:
				{
					fmt.Printf("client %s is already authenticated\n", simpConn.Id)
					break
				}
			}
		}
	}
}

func (broker *SimpBroker) Close() {
	if broker.Running {
		close(broker.serverClosingEvent)
	}
}
