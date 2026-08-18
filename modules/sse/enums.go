package sse

type EventType string

const (
	EventUpdateHistoryBalance EventType = "update_history_balance"
	EventOrder                EventType = "order"
)

func (e EventType) String() string {
	return string(e)
}

var eventTypeMap = map[string]EventType{
	string(EventUpdateHistoryBalance): EventUpdateHistoryBalance,
	string(EventOrder):                EventOrder,
}
