package pbft

import (
	"fmt"
	"math/rand"
	"time"
)

type StandardResponse struct {
	Detail string `json:"detail"`
}

func JsonBodyPadding(message string) string {
	body := fmt.Sprintf(`{"detail": "%v"}`, message)
	return body
}

func RandomNode(nodes []Node) Node {
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(nodes)
	return nodes[n]
}
