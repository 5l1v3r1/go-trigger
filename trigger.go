package trigger

import (
	"errors"
	"reflect"
	"sync"
)

func New() Trigger {
	return &trigger{
		events: make(map[string]*event),
	}
}

type trigger struct {
	events map[string]*event
	mu     sync.Mutex
}

type event struct {
	functions []*triggerfunc
	mu        sync.Mutex
}

type triggerfunc struct {
	function interface{}
}

var ForgottenFunction = errors.New("forgotten function")

func (t *trigger) Event(name string) Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	if e, found := t.events[name]; found {
		return e
	}
	e := event{}
	t.events[name] = &e
	return &e
}

// Backwards compatible
func (t *trigger) On(name string, task interface{}) error {
	_, err := t.Event(name).On(task)
	return err
}

func (e *event) On(task interface{}) (*triggerfunc, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if reflect.ValueOf(task).Type().Kind() != reflect.Func {
		return nil, errors.New("task is not a function")
	}
	tf := triggerfunc{function: task}
	e.functions = append(e.functions, &tf)
	return &tf, nil
}

func (e *event) Fire(params ...interface{}) ([][]reflect.Value, error) {
	var results [][]reflect.Value
	var ferr error
	for _, ef := range e.functions {
		f, in, err := ef.prepare(params...)
		if err != nil {
			if err != ForgottenFunction && ferr == nil {
				ferr = err
			}
		} else {
			results = append(results, f.Call(in))
		}
	}
	return results, ferr
}

func (e *event) FireBackground(params ...interface{}) (chan []reflect.Value, error) {
	results := make(chan []reflect.Value, len(e.functions))
	var wg sync.WaitGroup
	wg.Add(len(e.functions))
	go func() {
		for _, ef := range e.functions {
			f, in, err := ef.prepare(params...)
			if err == nil {
				// Background it
				go func() {
					results <- f.Call(in)
					wg.Done()
				}()
			} else {
				// Failed, but done
				wg.Done()
			}
		}
		wg.Wait()
		close(results)
	}()
	return results, nil
}

func (e *event) FireAndForget(params ...interface{}) error {
	go func() {
		for _, ef := range e.functions {
			f, in, err := ef.prepare(params...)
			if err == nil {
				f.Call(in)
			}
		}
	}()
	return nil
}

func (t *trigger) Clear(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.events[name]; !ok {
		return errors.New("event not defined")
	}
	delete(t.events, name)
	return nil
}

func (t *trigger) ClearEvents() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = make(map[string]*event)
}

func (t *trigger) HasEvent(name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.events[name]
	return ok
}

func (t *trigger) Events() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	events := make([]string, 0)
	for k := range t.events {
		events = append(events, k)
	}
	return events
}

func (t *trigger) EventCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.events)
}

func (ef *triggerfunc) Forget() error {
	ef.function = nil
	return nil
}

func (ef *triggerfunc) prepare(params ...interface{}) (reflect.Value, []reflect.Value, error) {
	if ef.function == nil {
		return reflect.Value{}, nil, ForgottenFunction
	}
	f := reflect.ValueOf(ef.function)
	if len(params) != f.Type().NumIn() {
		return reflect.Value{}, nil, errors.New("parameter mismatched")
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	return f, in, nil
}
