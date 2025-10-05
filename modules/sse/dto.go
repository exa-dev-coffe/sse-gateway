package sse

type BaseBodyMessage struct {
	Event  EventType `json:"event"`
	UserId int64     `json:"userId"`
}
