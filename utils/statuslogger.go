package utils

import (
	"fmt"
	"time"
)

const (
	INFO = iota
	WORK
	ERROR
)

type Message struct {
	Time time.Time
	Text string
	Type int8
}

var listeners chan Message
var messages [32]*Message

func Status(text string, _type int8) {
	for i := 1; i < len(messages); i++ {
		messages[i-1] = messages[i]
	}
}

func StatusInfof(format string, a ...interface{}) {
	Status(fmt.Sprintf(format, a...), INFO)
}
