package routers

import (
	"fmt"
	"key-value/lib/ws"
	"errors"
)

type Router interface {
	SetSetter(s RequestStrategy)
	SetGetter(s RequestStrategy)
	SetLister(s RequestStrategy)
	SetRemover(s RequestStrategy)
	CreateWebSocketHandler() ws.Handler
}

type router struct {
	getter  RequestStrategy
	setter  RequestStrategy
	lister  RequestStrategy
	remover RequestStrategy
}

func (r *router) SetSetter(s RequestStrategy) {
	r.setter = s
}

func (r *router) SetGetter(s RequestStrategy) {
	r.getter = s
}

func (r *router) SetLister(s RequestStrategy) {
	r.lister = s
}

func (r *router) SetRemover(s RequestStrategy) {
	r.remover = s
}

func (r *router) getActionStrategy(action string) RequestStrategy {
	if action == GET_DATA {
		return r.getter
	} else if action == SET_DATA {
		return r.setter
	} else if action == LIST_DATA {
		return r.lister
	} else if action == REMOVE_DATA {
		return r.remover
	}

	return func(r Request) (string, error) {
		return ``, errors.New(fmt.Sprintf(`Unexpected action: %d`, r.Action))
	}
}

func (r *router) CreateWebSocketHandler() ws.Handler {
	requestProcessor := func(request Request) (string, error) {
		strategy := r.getActionStrategy(request.Action)
		return strategy(request)
	}

	return createMessageHandler(createRequestHandler(requestProcessor))
}

func createDefaultStrategy(name string) RequestStrategy {
	return func(r Request) (string, error) {
		return ``, errors.New(name + ` not supported`)
	}
}

func New() Router {
	return &router{
		getter: createDefaultStrategy(GET_DATA),
		setter: createDefaultStrategy(SET_DATA),
		lister: createDefaultStrategy(LIST_DATA),
		remover: createDefaultStrategy(REMOVE_DATA),
	}
}