package rxgo

type ClosedObserverError struct {
}

// Observer represents a group of EventHandlers.
type Observer interface {
	EventHandler
	Disposable

	OnNext(item interface{}) error
	OnError(err error) error
	OnDone() error

	Block()
	setItemChannel(chan interface{})
	getItemChannel() chan interface{}
}

type observer struct {
	// itemChannel is the internal channel used to receive items from the parent observable
	itemChannel chan interface{}
	// nextHandler is the handler for the next items
	nextHandler NextFunc
	// errHandler is the error handler
	errHandler ErrFunc
	// doneHandler is the handler once an observable is done
	doneHandler DoneFunc
	// disposedChannel is the notification channel used when an observer is disposed
	disposedChannel chan struct{}
}

func (c *ClosedObserverError) Error() string {
	return "closed observer"
}

func (o *observer) setItemChannel(ch chan interface{}) {
	o.itemChannel = ch
}

func (o *observer) getItemChannel() chan interface{} {
	return o.itemChannel
}

// NewObserver constructs a new Observer instance with default Observer and accept
// any number of EventHandler
func NewObserver(eventHandlers ...EventHandler) Observer {
	ob := observer{
		disposedChannel: make(chan struct{}),
	}

	if len(eventHandlers) > 0 {
		for _, handler := range eventHandlers {
			switch handler := handler.(type) {
			case NextFunc:
				ob.nextHandler = handler
			case ErrFunc:
				ob.errHandler = handler
			case DoneFunc:
				ob.doneHandler = handler
			case *observer:
				ob = *handler
			}
		}
	}

	if ob.nextHandler == nil {
		ob.nextHandler = func(interface{}) {}
	}
	if ob.errHandler == nil {
		ob.errHandler = func(err error) {}
	}
	if ob.doneHandler == nil {
		ob.doneHandler = func() {}
	}

	return &ob
}

// Handle registers Observer to EventHandler.
func (o *observer) Handle(item interface{}) {
	switch item := item.(type) {
	default:
		o.nextHandler(item)
	case error:
		o.errHandler(item)
	}
}

func (o *observer) Dispose() {
	close(o.disposedChannel)
}

func (o *observer) Notify(ch chan<- struct{}) {
	ch <- struct{}{}
}

func (o *observer) IsDisposed() bool {
	select {
	case <-o.disposedChannel:
		return true
	default:
		return false
	}
}

// OnNext applies Observer's NextHandler to an Item
func (o *observer) OnNext(item interface{}) error {
	if !o.IsDisposed() {
		o.nextHandler(item)
		return nil
	} else {
		return &ClosedObserverError{}
	}
}

// OnError applies Observer's ErrHandler to an error
func (o *observer) OnError(err error) error {
	if !o.IsDisposed() {
		o.errHandler(err)
		o.Dispose()
		return nil
	} else {
		return &ClosedObserverError{}
	}
}

// OnDone terminates the Observer's internal Observable
func (o *observer) OnDone() error {
	if !o.IsDisposed() {
		o.doneHandler()
		o.Dispose()
		return nil
	} else {
		return &ClosedObserverError{}
	}
}

// OnDone terminates the Observer's internal Observable
func (o *observer) Block() {
	select {
	case <-o.disposedChannel:
		return
	}
}
