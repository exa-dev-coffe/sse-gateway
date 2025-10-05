package sse

type EventType string

const (
	EventUpdateHistoryBalance EventType = "update_history_balance"
)

func (e EventType) String() string {
	return string(e)
}

var eventTypeMap = map[string]EventType{
	string(EventUpdateHistoryBalance): EventUpdateHistoryBalance,
}
