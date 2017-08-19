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
			RequestId: request.RequestId,
		}
	}
}

func createMessageHandler(handler requestHandler) ws.Handler {
	return func(message []byte, sendQueue chan []byte) {
		var request Request
		err := json.Unmarshal(message, &request)
		if err != nil {
			msg := fmt.Sprintf(`Message: '%s' parse failed: %s`, message, err.Error())
			log.Println(msg)
			sendQueue <- []byte(msg)
			return
		}

		response := handler(request)

		responseJson, err := json.Marshal(response)
		if err != nil {
			log.Println(fmt.Sprintf(`Message: '%s' parse failed: %s`, message, err.Error()))
			return
		}

		sendQueue <- responseJson
	}
}