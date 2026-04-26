package godocker

import "encoding/json" // json.Unmarshal — converte o JSON da resposta em struct Go

// ══════════════════════════════════════════════════════════════════════════════
// FUNÇÕES DE PACOTE (atalho via DefaultClient)
// ══════════════════════════════════════════════════════════════════════════════

// Info retorna as informações gerais do daemon Docker usando o DefaultClient.
// Equivale a executar `docker info` no terminal.
// Inclui: versão, sistema operacional, CPUs, memória, plugins, swarm, etc.
func Info() (*DockerInfo, error) {
	return DefaultClient.Info()
}

// ══════════════════════════════════════════════════════════════════════════════
// MÉTODO NO CLIENT
// ══════════════════════════════════════════════════════════════════════════════

// Info retorna as informações completas do daemon Docker.
// Chama GET /info — este é geralmente o primeiro endpoint a ser testado
// para verificar se a conexão com o Docker está funcionando.
func (c *Client) Info() (*DockerInfo, error) {
	data, err := c.get("/info")
	if err != nil {
		// erro aqui geralmente significa que o Docker não está rodando
		// ou que o client não conseguiu conectar (socket/TCP inacessível)
		return nil, err
	}
	var out DockerInfo
	return &out, json.Unmarshal(data, &out)
}
