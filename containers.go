package godocker

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ── Funções de pacote (usam DefaultClient) ────────────────────────────────────

// ListContainers lista containers. all=true inclui os parados.
func ListContainers(all bool) ([]Container, error) {
	return DefaultClient.ListContainers(all)
}

// InspectContainer retorna detalhes completos de um container (ID ou nome).
func InspectContainer(id string) (*ContainerInspect, error) {
	return DefaultClient.InspectContainer(id)
}

// ContainerStats retorna um snapshot de uso de CPU e memória.
func ContainerStats(id string) (*Stats, error) {
	return DefaultClient.ContainerStats(id)
}

// ContainerLogs retorna as últimas `tail` linhas de log (stdout + stderr).
func ContainerLogs(id string, tail int) (string, error) {
	return DefaultClient.ContainerLogs(id, tail)
}

// CPUPercent calcula a porcentagem de CPU a partir de um Stats.
func CPUPercent(s *Stats) float64 {
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemCPUUsage) - float64(s.PreCPUStats.SystemCPUUsage)
	cpus := float64(s.CPUStats.OnlineCPUs)
	if cpus == 0 {
		cpus = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
	}
	if sysDelta == 0 {
		return 0
	}
	return (cpuDelta / sysDelta) * cpus * 100.0
}

// ── Métodos no Client ─────────────────────────────────────────────────────────

// ListContainers lista containers. all=true inclui os parados.
func (c *Client) ListContainers(all bool) ([]Container, error) {
	path := "/containers/json"
	if all {
		path += "?all=true"
	}
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var out []Container
	return out, json.Unmarshal(data, &out)
}

// InspectContainer retorna detalhes completos de um container (ID ou nome).
func (c *Client) InspectContainer(id string) (*ContainerInspect, error) {
	data, err := c.get("/containers/" + id + "/json")
	if err != nil {
		return nil, err
	}
	var out ContainerInspect
	return &out, json.Unmarshal(data, &out)
}

// ContainerStats retorna um snapshot de uso de CPU e memória (stream=false).
func (c *Client) ContainerStats(id string) (*Stats, error) {
	data, err := c.get("/containers/" + id + "/stats?stream=false")
	if err != nil {
		return nil, err
	}
	var out Stats
	return &out, json.Unmarshal(data, &out)
}

// ContainerLogs retorna as últimas `tail` linhas de log (stdout + stderr).
// Faz o parse do multiplexed stream do Docker:
// cada frame tem header [type(1), 0,0,0, size(4 BE)] seguido do payload.
func (c *Client) ContainerLogs(id string, tail int) (string, error) {
	path := fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true&tail=%d", id, tail)
	data, err := c.get(path)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	b := data
	for len(b) > 8 {
		size := int(b[4])<<24 | int(b[5])<<16 | int(b[6])<<8 | int(b[7])
		if 8+size > len(b) {
			break
		}
		sb.Write(b[8 : 8+size])
		b = b[8+size:]
	}
	return sb.String(), nil
}