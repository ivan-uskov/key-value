package routers

const (
	GET    = `GET`
	SET    = `SET`
	LIST   = `LIST`
	REMOVE = `REMOVE`
	PING   = `PING`
)

type Request struct {
	Action string
	Option1 string
	Option2 string
}

type response struct {
	Success bool
	Error string
	Result string
}

type requestHandler func (request Request) response
type RequestStrategy func(r Request) (string, error)