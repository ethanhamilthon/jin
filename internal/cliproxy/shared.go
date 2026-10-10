package cliproxy

import "sync"

var shared struct {
	sync.Mutex
	clients map[string]*Client
}

func Shared(root string) *Client {
	shared.Lock()
	defer shared.Unlock()
	if shared.clients == nil {
		shared.clients = map[string]*Client{}
	}
	if shared.clients[root] == nil {
		shared.clients[root] = NewClient(root)
	}
	return shared.clients[root]
}

func CloseAll() {
	shared.Lock()
	defer shared.Unlock()
	for root, client := range shared.clients {
		client.Close()
		delete(shared.clients, root)
	}
}
