package agent

var baseURL = "http://localhost:8080"

type Client interface {
	HeartBeat() error
	Submit(string, interface{}) error
}

type CanaryClient struct {
	apiKey     string
	serverName string
}

func NewClient(apiKey string, serverName string) *CanaryClient {
	client := &CanaryClient{apiKey: apiKey, serverName: serverName}
	return client
}

func (c *CanaryClient) HeartBeat() error {
	//TODO
	return nil
}

func (c *CanaryClient) Submit(app string, data interface{}) error {
	//TODO
	return nil
}
