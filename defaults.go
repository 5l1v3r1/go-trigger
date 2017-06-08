package trigger

import "reflect"

type Trigger interface {
	On(name string, task interface{}) error
	Event(name string) Event
	Fire(name string, params ...interface{}) ([][]reflect.Value, error)
	FireBackground(name string, params ...interface{}) (chan []reflect.Value, error)
	Clear(name string) error
	ClearEvents()
	HasEvent(name string) bool
	Events() []string
	EventCount() int
}

type Event interface {
	On(task interface{}) (*triggerfunc, error)
	Fire(params ...interface{}) ([][]reflect.Value, error)
	FireBackground(params ...interface{}) (chan []reflect.Value, error)
	FireAndForget(params ...interface{}) error
}

type Function interface {
	Forget()
}

var defaultTrigger = New()

func Global() Trigger {
	return defaultTrigger
}

// Default global trigger options.
func On(name string, task interface{}) error {
	_, err := defaultTrigger.Event(name).On(task)
	return err
}

func Fire(name string, params ...interface{}) ([][]reflect.Value, error) {
	return defaultTrigger.Event(name).Fire(params...)
}

func FireBackground(name string, params ...interface{}) (chan []reflect.Value, error) {
	return defaultTrigger.Event(name).FireBackground(params...)
}

func Clear(event string) error {
	return defaultTrigger.Clear(event)
}

func ClearEvents() {
	defaultTrigger.ClearEvents()
}

func HasEvent(event string) bool {
	return defaultTrigger.HasEvent(event)
}

func Events() []string {
	return defaultTrigger.Events()
}

func EventCount() int {
	return defaultTrigger.EventCount()
}
