package routers

const (
	GET_DATA = `GET`
	SET_DATA = `SET`
	LIST_DATA = `LIST`
	REMOVE_DATA = `REMOVE`
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