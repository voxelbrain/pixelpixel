package main

type MessageType int

const (
	TypeCreate MessageType = iota
	TypeChange
	TypeFailure
	TypeDelete
)

type Message struct {
	Pixel   string      `json:"pixel"`
	Type    MessageType `json:"type"`
	Payload string      `json:"payload"`
}
