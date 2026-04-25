package godocker

import "encoding/json"

// ListVolumes lista os volumes usando o DefaultClient.
func ListVolumes() ([]Volume, error) {
	return DefaultClient.ListVolumes()
}

// ListVolumes lista os volumes Docker.
func (c *Client) ListVolumes() ([]Volume, error) {
	data, err := c.get("/volumes")
	if err != nil {
		return nil, err
	}
	var result struct {
		Volumes []Volume `json:"Volumes"`
	}
	return result.Volumes, json.Unmarshal(data, &result)
}