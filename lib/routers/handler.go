package routers

import (
	"key-value/lib/ws"
	"encoding/json"
	"log"
	"fmt"
)

func createRequestHandler(strategy RequestStrategy) requestHandler {
	return func (request Request) response {
		value, err := strategy(request)
		errorMsg := ``
		if err != nil {
			errorMsg = err.Error()
		}

		return response{
			Success:   err == nil,
			Error:     errorMsg,
			Result:    value,
		}
	}
}

func createMessageHandler(handler requestHandler) ws.RequestHandler {
	return func(message []byte) []byte {
		var request Request
		err := json.Unmarshal(message, &request)
		if err != nil {
			msg := fmt.Sprintf(`Message: '%s' parse failed: %s`, message, err.Error())
			log.Println(msg)
			return []byte(msg)
		}

		response := handler(request)

		responseJson, err := json.Marshal(response)
		if err != nil {
			msg := fmt.Sprintf(`Message: '%s' encode response failed: %s`, message, err.Error())
			log.Println(msg)
			return []byte(msg)
		}

		return responseJson
	}
}