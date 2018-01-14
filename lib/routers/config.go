package routers

const (
	GET    = `GET`
	SET    = `SET`
	LIST   = `LIST`
	REMOVE = `REMOVE`
	PING   = `PING`
	RUN    = `RUN`
)

type Request struct {
	Action  string `json:"action"`
	Option1 string `json:"option_1"`
	Option2 string `json:"option_2"`
}

type Response struct {
	Success bool `json:"success"`
	Error   string `json:"error"`
	Result  string `json:"result"`
}

type requestHandler func(request Request) Response
type RequestStrategy func(r Request) (string, error)
