package main

type MessageType int

const (
	TypeCreate MessageType = iota
	TypeChange
	TypeFailure
)

type Message struct {
	Pixel   string      `json:"pixel"`
	Type    MessageType `json:"type"`
	Payload string      `json:"payload"`
}
