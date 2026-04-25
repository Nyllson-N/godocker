package godocker

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ── Funções de pacote (usam DefaultClient) ────────────────────────────────────

func ListNetworks() ([]Network, error)                                              { return DefaultClient.ListNetworks() }
func InspectNetwork(id string) (*Network, error)                                    { return DefaultClient.InspectNetwork(id) }
func CreateNetwork(opts NetworkCreateOptions) (string, error)                       { return DefaultClient.CreateNetwork(opts) }
func RemoveNetwork(id string) error                                                 { return DefaultClient.RemoveNetwork(id) }
func ConnectNetwork(networkID, containerID string, opts *NetworkConnectOptions) error { return DefaultClient.ConnectNetwork(networkID, containerID, opts) }
func DisconnectNetwork(networkID, containerID string, force bool) error             { return DefaultClient.DisconnectNetwork(networkID, containerID, force) }
func PruneNetworks() ([]string, error)                                              { return DefaultClient.PruneNetworks() }
func NetworkExists(nameOrID string) (bool, error)                                   { return DefaultClient.NetworkExists(nameOrID) }
func RenameNetwork(oldName, newName string) (string, error)                         { return DefaultClient.RenameNetwork(oldName, newName) }

// ── Métodos no Client ─────────────────────────────────────────────────────────

// ListNetworks retorna todas as redes Docker.
func (c *Client) ListNetworks() ([]Network, error) {
	data, err := c.get("/networks")
	if err != nil {
		return nil, err
	}
	var out []Network
	return out, json.Unmarshal(data, &out)
}

// InspectNetwork retorna detalhes de uma rede (ID ou nome),
// incluindo os containers conectados.
func (c *Client) InspectNetwork(id string) (*Network, error) {
	data, err := c.get("/networks/" + id + "?verbose=true")
	if err != nil {
		return nil, err
	}
	var out Network
	return &out, json.Unmarshal(data, &out)
}

// CreateNetwork cria uma nova rede e retorna o ID gerado.
// Driver padrão é "bridge" se omitido.
//
// Exemplo mínimo:
//
//	id, err := godocker.CreateNetwork(godocker.NetworkCreateOptions{
//	    Name: "minha-rede",
//	})
//
// Exemplo com subnet customizada:
//
//	id, err := godocker.CreateNetwork(godocker.NetworkCreateOptions{
//	    Name:   "minha-rede",
//	    Driver: "bridge",
//	    IPAM: &godocker.NetworkIPAM{
//	        Driver: "default",
//	        Config: []godocker.IPAMConfig{
//	            {Subnet: "172.28.0.0/16", Gateway: "172.28.0.1"},
//	        },
//	    },
//	})
func (c *Client) CreateNetwork(opts NetworkCreateOptions) (string, error) {
	if opts.Driver == "" {
		opts.Driver = "bridge"
	}
	data, err := c.post("/networks/create", opts)
	if err != nil {
		return "", err
	}
	var result struct {
		ID      string `json:"Id"`
		Warning string `json:"Warning"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if result.Warning != "" {
		fmt.Printf("⚠️  Docker warning: %s\n", result.Warning)
	}
	return result.ID, nil
}

// RemoveNetwork remove uma rede pelo ID ou nome.
// A rede não pode ter containers ativos conectados.
func (c *Client) RemoveNetwork(id string) error {
	_, err := c.delete("/networks/" + id)
	return err
}

// ConnectNetwork conecta um container a uma rede existente.
// Passe nil em opts para usar IP automático sem alias.
//
// Exemplo com IP fixo e alias:
//
//	err := godocker.ConnectNetwork("minha-rede", "meu-container", &godocker.NetworkConnectOptions{
//	    EndpointConfig: &godocker.EndpointConfig{
//	        IPAMConfig: &godocker.EndpointIPAMConfig{IPv4Address: "172.28.0.10"},
//	        Aliases:    []string{"backend"},
//	    },
//	})
func (c *Client) ConnectNetwork(networkID, containerID string, opts *NetworkConnectOptions) error {
	if opts == nil {
		opts = &NetworkConnectOptions{}
	}
	opts.Container = containerID
	_, err := c.post("/networks/"+networkID+"/connect", opts)
	return err
}

// DisconnectNetwork desconecta um container de uma rede.
// force=true desconecta mesmo que o container esteja rodando.
func (c *Client) DisconnectNetwork(networkID, containerID string, force bool) error {
	body := map[string]any{
		"Container": containerID,
		"Force":     force,
	}
	_, err := c.post("/networks/"+networkID+"/disconnect", body)
	return err
}

// PruneNetworks remove todas as redes sem containers conectados.
// Retorna os nomes das redes removidas.
func (c *Client) PruneNetworks() ([]string, error) {
	data, err := c.post("/networks/prune", nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		NetworksDeleted []string `json:"NetworksDeleted"`
	}
	return result.NetworksDeleted, json.Unmarshal(data, &result)
}

// NetworkExists verifica se uma rede com o nome ou ID fornecido existe.
func (c *Client) NetworkExists(nameOrID string) (bool, error) {
	_, err := c.InspectNetwork(nameOrID)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// RenameNetwork recria a rede com um novo nome preservando driver, IPAM e labels.
// A Docker API não possui endpoint nativo de rename para redes.
// ATENÇÃO: containers conectados precisam ser reconectados manualmente.
func (c *Client) RenameNetwork(oldName, newName string) (string, error) {
	old, err := c.InspectNetwork(oldName)
	if err != nil {
		return "", fmt.Errorf("rede '%s' não encontrada: %w", oldName, err)
	}

	newID, err := c.CreateNetwork(NetworkCreateOptions{
		Name:       newName,
		Driver:     old.Driver,
		Internal:   old.Internal,
		Attachable: old.Attachable,
		EnableIPv6: old.EnableIPv6,
		Options:    old.Options,
		Labels:     old.Labels,
		IPAM:       &old.IPAM,
	})
	if err != nil {
		return "", fmt.Errorf("falha ao criar rede '%s': %w", newName, err)
	}

	if err := c.RemoveNetwork(oldName); err != nil {
		return newID, fmt.Errorf("nova rede criada (%s) mas falha ao remover a antiga: %w", newID, err)
	}
	return newID, nil
}