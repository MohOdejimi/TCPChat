package models

import "time"

type Message struct {
	Sender string 
	Message string
	Time time.Time 
}

type DMMessage struct {
	Sender string 
	Receiver string 
	Message string 
	Time time.Time
}