package godocker

import (
	"encoding/json" // json.Unmarshal / json.Marshal — conversão entre JSON e structs Go
	"fmt"           // fmt.Errorf — formata mensagens de erro com contexto
	"strings"       // strings.Contains — verifica se o erro contém "404" (rede não encontrada)
)

// ══════════════════════════════════════════════════════════════════════════════
// FUNÇÕES DE PACOTE (atalhos via DefaultClient)
// ══════════════════════════════════════════════════════════════════════════════

// ListNetworks lista todas as redes Docker disponíveis na máquina.
// Equivale a `docker network ls`.
func ListNetworks() ([]Network, error) { return DefaultClient.ListNetworks() }

// InspectNetwork retorna os detalhes completos de uma rede pelo ID ou nome.
// Inclui containers conectados, configuração IPAM e opções do driver.
func InspectNetwork(id string) (*Network, error) { return DefaultClient.InspectNetwork(id) }

// CreateNetwork cria uma nova rede Docker e retorna o ID gerado.
// O driver padrão é "bridge" se não for especificado em opts.
func CreateNetwork(opts NetworkCreateOptions) (string, error) {
	return DefaultClient.CreateNetwork(opts)
}

// RemoveNetwork remove uma rede pelo ID ou nome.
// A rede não pode ter containers ativos conectados no momento da remoção.
func RemoveNetwork(id string) error { return DefaultClient.RemoveNetwork(id) }

// ConnectNetwork conecta um container existente a uma rede existente.
// Passe nil em opts para usar IP automático sem aliases.
func ConnectNetwork(networkID, containerID string, opts *NetworkConnectOptions) error {
	return DefaultClient.ConnectNetwork(networkID, containerID, opts)
}

// DisconnectNetwork desconecta um container de uma rede.
// force=true desconecta mesmo que o container esteja em execução.
func DisconnectNetwork(networkID, containerID string, force bool) error {
	return DefaultClient.DisconnectNetwork(networkID, containerID, force)
}

// PruneNetworks remove todas as redes que não têm nenhum container conectado.
// Retorna os nomes das redes que foram removidas.
func PruneNetworks() ([]string, error) { return DefaultClient.PruneNetworks() }

// NetworkExists verifica se uma rede com o nome ou ID fornecido existe.
// Retorna true se existir, false se não existir (sem retornar erro no caso 404).
func NetworkExists(nameOrID string) (bool, error) { return DefaultClient.NetworkExists(nameOrID) }

// RenameNetwork recria a rede com um novo nome preservando suas configurações.
// Atenção: a Docker API não tem endpoint nativo de rename para redes,
// então esta função cria uma nova rede e remove a antiga.
// Containers conectados perdem a conexão e precisam ser reconectados.
func RenameNetwork(oldName, newName string) (string, error) {
	return DefaultClient.RenameNetwork(oldName, newName)
}

// ══════════════════════════════════════════════════════════════════════════════
// MÉTODOS NO CLIENT
// ══════════════════════════════════════════════════════════════════════════════

// ListNetworks lista todas as redes Docker chamando GET /networks.
func (c *Client) ListNetworks() ([]Network, error) {
	data, err := c.get("/networks")
	if err != nil {
		return nil, err
	}
	var out []Network
	return out, json.Unmarshal(data, &out)
}

// InspectNetwork retorna os detalhes de uma rede específica.
// O parâmetro verbose=true faz o Docker incluir os containers conectados no resultado.
func (c *Client) InspectNetwork(id string) (*Network, error) {
	data, err := c.get("/networks/" + id + "?verbose=true")
	if err != nil {
		return nil, err
	}
	var out Network
	return &out, json.Unmarshal(data, &out)
}

// CreateNetwork cria uma nova rede Docker e retorna o ID gerado.
// Chama POST /networks/create com as opções serializadas como JSON.
// Se o driver não for especificado, usa "bridge" como padrão.
func (c *Client) CreateNetwork(opts NetworkCreateOptions) (string, error) {
	if opts.Driver == "" {
		opts.Driver = "bridge" // driver padrão do Docker para redes locais
	}
	data, err := c.post("/networks/create", opts)
	if err != nil {
		return "", err
	}
	// o Docker retorna {"Id": "abc123...", "Warning": ""}
	var result struct {
		ID      string `json:"Id"`
		Warning string `json:"Warning"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	// avisos não são erros fatais — o Docker criou a rede mas tem algo a informar
	if result.Warning != "" {
		fmt.Printf("aviso Docker: %s\n", result.Warning)
	}
	return result.ID, nil
}

// RemoveNetwork remove uma rede chamando DELETE /networks/{id}.
func (c *Client) RemoveNetwork(id string) error {
	_, err := c.delete("/networks/" + id)
	return err
}

// ConnectNetwork conecta um container a uma rede existente.
// Chama POST /networks/{networkID}/connect com o ID do container e as opções.
// Se opts for nil, cria um objeto vazio — o Docker atribui IP automático.
func (c *Client) ConnectNetwork(networkID, containerID string, opts *NetworkConnectOptions) error {
	if opts == nil {
		opts = &NetworkConnectOptions{} // sem configuração específica → IP automático
	}
	opts.Container = containerID // o Docker precisa do ID do container no corpo do request
	_, err := c.post("/networks/"+networkID+"/connect", opts)
	return err
}

// DisconnectNetwork desconecta um container de uma rede.
// Chama POST /networks/{networkID}/disconnect — é um POST, não um DELETE,
// porque o Docker precisa receber o ID do container e o flag force no corpo.
func (c *Client) DisconnectNetwork(networkID, containerID string, force bool) error {
	body := map[string]any{
		"Container": containerID,
		"Force":     force, // true = força desconexão mesmo com container rodando
	}
	_, err := c.post("/networks/"+networkID+"/disconnect", body)
	return err
}

// PruneNetworks remove redes sem containers conectados.
// Chama POST /networks/prune — o Docker retorna {"NetworksDeleted": ["rede1", "rede2"]}.
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

// NetworkExists verifica se uma rede existe tentando fazer um Inspect.
// Se o Inspect retornar erro com "404", a rede não existe (false, nil).
// Outros erros são propagados normalmente.
func (c *Client) NetworkExists(nameOrID string) (bool, error) {
	_, err := c.InspectNetwork(nameOrID)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil // 404 = não encontrado → retorna false sem erro
		}
		return false, err // outro erro (rede, permissão, etc.) → propaga o erro
	}
	return true, nil
}

// RenameNetwork recria a rede com um novo nome preservando driver, IPAM e labels.
//
// Por que recria em vez de renomear?
// A Docker Engine API não possui um endpoint de rename para redes (diferente de containers).
// A solução é: fazer inspect da rede antiga → criar nova rede com mesmo config e novo nome
// → remover a rede antiga.
//
// Limitação importante: containers que estavam conectados à rede antiga perdem a conexão
// e precisam ser reconectados manualmente à nova rede.
func (c *Client) RenameNetwork(oldName, newName string) (string, error) {
	// busca as configurações atuais da rede antiga
	old, err := c.InspectNetwork(oldName)
	if err != nil {
		return "", fmt.Errorf("rede '%s' não encontrada: %w", oldName, err)
	}

	// cria a nova rede com as mesmas configurações mas com o novo nome
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

	// remove a rede antiga — se falhar, a nova rede já foi criada
	// retornamos o ID da nova rede junto com o erro para o chamador poder limpar
	if err := c.RemoveNetwork(oldName); err != nil {
		return newID, fmt.Errorf("nova rede criada (%s) mas falha ao remover a antiga: %w", newID, err)
	}
	return newID, nil
}
