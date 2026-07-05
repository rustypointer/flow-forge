package events

type Sink interface {
	Publish(Event)
}
