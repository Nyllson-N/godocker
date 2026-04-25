package godocker

import "encoding/json"

// ListImages lista as imagens locais usando o DefaultClient.
func ListImages() ([]Image, error) {
	return DefaultClient.ListImages()
}

// ListImages lista as imagens locais.
func (c *Client) ListImages() ([]Image, error) {
	data, err := c.get("/v1.41/images/json")
	if err != nil {
		return nil, err
	}
	var out []Image
	return out, json.Unmarshal(data, &out)
}
