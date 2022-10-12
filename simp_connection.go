package simpmq

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"yadunandan.xyz/simp_mq/slog"
)

//stores details of connection from a client on server
type SimpClientConn struct {
	//
	NetConn net.Conn //actual connection

	authenticated bool //whether this connection has been authenticated or not

	BufferSize uint //size of the each message, maintain uniformity across client and broker

	Authenticator Authenticator //authenticate connection using this callback

	WaitForAuthentication time.Duration //wait window till the AuthDetails arrives after which connection fails

	Id string //id
}

//stores connection to a server on client
type SimpServerConn struct {
	NetConn net.Conn //actual connection

	authenticated bool //whether this connection has been authenticated or not

	BufferSize uint //size of the each message, maintain uniformity across client and broker

	AuthDetails *AuthDetails //details to authenticate with broker

	Id string //id
}

//attempts to authenticate with the server using the AuthDetails
func (sc *SimpServerConn) authenticateWithBroker() (err error) {
	bytes, err := json.Marshal(sc.AuthDetails)
	if err != nil {
		return err
	}
	err = sc.respond(&SimpData{Type: auth, Payload: bytes, ID: sc.Id})

	if err != nil {
		return err
	}
	data, err := sc.nextDataFromConnection()
	if err != nil {
		return err
	}
	if data.Type != authAck || data.ID != sc.Id {
		return fmt.Errorf("failed to authenticate from client because no auth ack recieved")
	}
	sc.authenticated = true
	return nil
}

//fails if not authenticated, waits for next data to arrive
func (sc *SimpServerConn) nextDataFromConnection() (*SimpData, error) {
	if !sc.authenticated {
		return nil, fmt.Errorf("connection is not authenticated to read")
	}
	return nextDataFromConnection(sc.BufferSize, sc.NetConn)
}

//fails if not authenticated, send data to server
func (sc *SimpServerConn) respond(data *SimpData) (err error) {
	if !sc.authenticated {
		return fmt.Errorf("connection is not authenticated to respond")
	}
	return respond(data, sc.NetConn)
}

//authenticates using provided autheticator funtion provided to the instance
//return the auth data or else error
func (sc *SimpClientConn) authenticateWithClient() (data *SimpData, err error) {
	if sc.WaitForAuthentication == 0 {
		sc.WaitForAuthentication = time.Second * 16
	}
	data, err = sc.nextDataFromConnectionWithWait(sc.WaitForAuthentication)
	if err != nil {
		return nil, err
	}
	if sc.Authenticator != nil {
		deets, err := data.GetAuthDetails()
		if err != nil {
			return nil, err
		}
		err = sc.Authenticator(deets)
		if err != nil {
			return nil, err
		}
		if len(deets.ClientID) == 0 {
			return nil, errors.New("empty string cannot be clientId")
		} else {
			sc.Id = deets.ClientID
		}
		return data, nil
	} else {
		return nil, fmt.Errorf("simp broker Authenticator must be provided")
	}
}

//closes the connection
func (sc *SimpClientConn) close() {
	err := sc.NetConn.Close()
	if err != nil {
		slog.Warn("error closing connection simp_connection: close()")
	}
}

//fails if not authenticated, waits for next data to arrive
func (sc *SimpClientConn) nextDataFromConnection() (*SimpData, error) {
	if !sc.authenticated {
		return nil, fmt.Errorf("connection is not authenticated to read")
	}
	return nextDataFromConnection(sc.BufferSize, sc.NetConn)
}

//fails if not authenticated or time's up
//waits for next data to arrive
func (sc *SimpClientConn) nextDataFromConnectionWithWait(t time.Duration) (*SimpData, error) {
	if !sc.authenticated {
		return nil, fmt.Errorf("connection is not authenticated to read")
	}
	ch := make(chan *SimpData)
	var err error
	go func() {
		conn := sc.NetConn
		conn.SetReadDeadline(time.Now().Add(t))
		data, err := nextDataFromConnection(sc.BufferSize, sc.NetConn)
		conn.SetReadDeadline(time.Time{})
		if err == nil {
			ch <- data
		}
		close(ch)
	}()
	r1, _ := <-ch
	return r1, err
}

//send data to the client
func (sc *SimpClientConn) respond(data *SimpData) (err error) {
	if !sc.authenticated {
		return fmt.Errorf("connection is not authenticated to respond")
	}
	return respond(data, sc.NetConn)
}

func respond(data *SimpData, NetConn net.Conn) (err error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = NetConn.Write(bytes)
	if err != nil {
		return err
	}
	return
}

func nextDataFromConnection(BufferSize uint, NetConn net.Conn) (*SimpData, error) {
	buf := make([]byte, BufferSize)
	readLen, err := NetConn.Read(buf[0:])

	if err != nil {
		return nil, err

	} else {
		simpData := &SimpData{}
		err = json.Unmarshal(buf[:readLen], simpData)
		if err != nil {
			return nil, err
		} else {
			return simpData, nil
		}
	}
}
