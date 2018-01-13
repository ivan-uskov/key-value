package routers

import (
	"fmt"
	"key-value/lib/ws"
	"errors"
)

type Router interface {
	AddRoute(address string, s RequestStrategy)
	CreateWebSocketHandler() ws.RequestHandler
}

type router struct {
	routing map[string]RequestStrategy
}

func (r *router) AddRoute(address string, s RequestStrategy) {
	r.routing[address] = s
}

func (r *router) getActionStrategy(action string) RequestStrategy {
	s, ok := r.routing[action]
	if !ok {
		return func(r Request) (string, error) {
			return ``, errors.New(fmt.Sprintf(`Unexpected action: %d`, r.Action))
		}
	}

	return s
}

func (r *router) CreateWebSocketHandler() ws.RequestHandler {
	requestProcessor := func(request Request) (string, error) {
		strategy := r.getActionStrategy(request.Action)
		return strategy(request)
	}

	return createMessageHandler(createRequestHandler(requestProcessor))
}

func NewRouter() Router {
	return &router{routing: make(map[string]RequestStrategy)}
}