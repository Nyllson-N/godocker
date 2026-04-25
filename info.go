package godocker

import "encoding/json"

// Info retorna informações gerais do daemon Docker
// usando o DefaultClient.
func Info() (*DockerInfo, error) {
	return DefaultClient.Info()
}

// Info retorna informações gerais do daemon Docker.
func (c *Client) Info() (*DockerInfo, error) {
	data, err := c.get("/v1.41/info")
	if err != nil {
		return nil, err
	}
	var out DockerInfo
	return &out, json.Unmarshal(data, &out)
}
