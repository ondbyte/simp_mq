// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    simpData, err := UnmarshalSimpData(bytes)
//    bytes, err = simpData.Marshal()

package simp_broker

import (
	"encoding/json"
)

func UnmarshalSimpData(data []byte) (SimpData, error) {
	var r SimpData
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SimpData) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *SimpData) GetAuthDetails() (*AuthDetails, error) {
	return UnmarshalAuthDetails(r.Payload)
}

func (r *SimpData) GetSubDetails() (*SubDetails, error) {
	return UnmarshalSubDetails(r.Payload)
}

func (r *SimpData) GetPubDetails() (*PubDetails, error) {
	return UnmarshalPubDetails(r.Payload)
}

type SimpData struct {
	Type    MessagType `json:"type,omitempty"`
	ID      string     `json:"id,omitempty"`
	Payload []byte     `json:"payload,omitempty"`
}

func UnmarshalAuthDetails(data []byte) (*AuthDetails, error) {
	r := &AuthDetails{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AuthDetails) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type AuthDetails struct {
	Token    string `json:"token,omitempty"`
	ClientID string `json:"clientId,omitempty"`
}

type MessagType int

const (
	sub MessagType = iota
	subAck
	pub
	pubAck
	auth
	authAck
	unsub
	unsubAck
)

func UnmarshalSubDetails(data []byte) (*SubDetails, error) {
	r := &SubDetails{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SubDetails) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type SubDetails struct {
	Topic string `json:"topic,omitempty"`
}

func UnmarshalPubDetails(data []byte) (*PubDetails, error) {
	r := &PubDetails{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PubDetails) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type PubDetails struct {
	Topic string `json:"topic,omitempty"`
	Data  []byte `json:"data,omitempty"`
}

type Authenticator func(*AuthDetails) error
