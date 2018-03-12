package replication

import (
	"key-value/instance/storages"
	"net/http"
	"key-value/lib/routers"
	"key-value/lib/ws"
)

type Server interface {
	Bind()
}

type server struct {
	storage storages.Storage
	client Client
}

func NewServer(storage storages.Storage, client Client) Server {
	return &server{storage, client}
}

func (s *server) Bind() {
	r := s.createRouter()
	wsServer := ws.NewServer()
	http.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
		wsServer.Serve(writer, request, r.CreateWebSocketHandler())
	})
}

func (s *server) createRouter() routers.Router {
	r := routers.NewRouter()
	r.AddRoute(updated, func(r routers.Request) (string, error) {
		s.storage.SetWithVersion(r.Option1, r.Option2, r.Version)
		return ``, nil
	})

	r.AddRoute(removed, func(r routers.Request) (string, error) {
		s.storage.RemoveWithVersion(r.Option1, r.Version)
		return ``, nil
	})

	return r
}
