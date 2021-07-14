package pbft

import (
	"fmt"
)

type StandardResponse struct {
	Detail string `json:"detail"`
}

func HttpJsonBodyPadding(message string) string {
	body := fmt.Sprintf("{\"detail\": \"%v\"}", message)
	return body
}
