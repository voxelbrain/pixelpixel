package main

import (
	"bytes"
	"fmt"

	"github.com/voxelbrain/pixelpixel/pixelutils"
)

type Pixel struct {
	Container  `json:"-"`
	Id         string                 `json:"id"`
	Broken     bool                   `json:"broken"`
	Filesystem map[string]string      `json:"-"`
	Clicks     chan *pixelutils.Click `json:"-"`
	LastImage  *bytes.Buffer          `json:"-"`
}

func (p *Pixel) Fail(msgf string, data ...interface{}) *Message {
	p.Broken = true
	return &Message{
		Pixel:   p.Id,
		Type:    TypeFailure,
		Payload: fmt.Sprintf(msgf, data...),
	}
}
