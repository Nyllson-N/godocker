package godocker

import (
	"encoding/json" // json.Unmarshal — converte JSON bytes em structs Go
	"fmt"           // fmt.Sprintf — formata a URL com parâmetros
	"strings"       // strings.Builder — constrói a string dos logs de forma eficiente
)

// ══════════════════════════════════════════════════════════════════════════════
// FUNÇÕES DE PACOTE (atalhos que usam o DefaultClient)
// Permitem chamar docker.ListContainers() sem criar um Client manualmente.
// ══════════════════════════════════════════════════════════════════════════════

// ListContainers lista os containers do Docker.
// all=false → retorna apenas containers em execução (estado "running")
// all=true  → retorna todos os containers, incluindo os parados e com erro
func ListContainers(all bool) ([]Container, error) {
	return DefaultClient.ListContainers(all)
}

// InspectContainer retorna os detalhes completos de um container.
// O parâmetro id aceita tanto o ID completo quanto o nome do container.
// Exemplo: InspectContainer("meu-nginx") ou InspectContainer("a1b2c3d4e5f6")
func InspectContainer(id string) (*ContainerInspect, error) {
	return DefaultClient.InspectContainer(id)
}

// ContainerStats retorna um snapshot do uso de recursos de um container.
// Inclui: CPU, memória, I/O de disco e tráfego de rede.
// Só funciona para containers em execução (estado "running").
func ContainerStats(id string) (*Stats, error) {
	return DefaultClient.ContainerStats(id)
}

// ContainerLogs retorna as últimas `tail` linhas de log do container.
// Captura tanto stdout quanto stderr juntos em uma única string.
func ContainerLogs(id string, tail int) (string, error) {
	return DefaultClient.ContainerLogs(id, tail)
}

// CPUPercent calcula a porcentagem de uso de CPU a partir de um snapshot de Stats.
// O Docker fornece contadores cumulativos (nanossegundos), então precisamos de
// dois snapshots (atual e anterior) para calcular o delta de uso.
//
// Fórmula:
//
//	cpu% = (delta_uso_container / delta_uso_sistema) × número_de_CPUs × 100
func CPUPercent(s *Stats) float64 {
	// delta de uso do container entre a leitura atual e a anterior
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)

	// delta de uso do sistema inteiro (todos os processos) no mesmo período
	sysDelta := float64(s.CPUStats.SystemCPUUsage) - float64(s.PreCPUStats.SystemCPUUsage)

	// número de CPUs lógicas disponíveis para o container
	cpus := float64(s.CPUStats.OnlineCPUs)
	if cpus == 0 {
		// fallback: conta pelo tamanho da lista percpu (versões mais antigas do Docker)
		cpus = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
	}

	// evita divisão por zero quando os dois snapshots são idênticos
	if sysDelta == 0 {
		return 0
	}

	return (cpuDelta / sysDelta) * cpus * 100.0
}

// ══════════════════════════════════════════════════════════════════════════════
// MÉTODOS NO CLIENT
// Implementação real — as funções de pacote acima apenas delegam para cá.
// ══════════════════════════════════════════════════════════════════════════════

// ListContainers lista os containers do Docker.
// Chama o endpoint GET /containers/json da Docker Engine API.
// Com all=true adiciona o parâmetro ?all=true na URL.
func (c *Client) ListContainers(all bool) ([]Container, error) {
	path := "/containers/json?all=true" // sempre busca todos
	if !all {
		path = "/containers/json" // só running quando explicitamente pedido
	}
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var out []Container
	return out, json.Unmarshal(data, &out)
}

// InspectContainer retorna os detalhes completos de um container.
// Chama GET /containers/{id}/json — retorna muito mais informações que ListContainers,
// incluindo configuração de rede, volumes, variáveis de ambiente, healthcheck, etc.
func (c *Client) InspectContainer(id string) (*ContainerInspect, error) {
	data, err := c.get("/containers/" + id + "/json")
	if err != nil {
		return nil, err
	}
	var out ContainerInspect
	return &out, json.Unmarshal(data, &out)
}

// ContainerStats retorna um snapshot do uso de recursos do container.
// Usa stream=false para receber apenas UMA leitura e fechar a conexão.
// Sem esse parâmetro o Docker ficaria enviando dados continuamente (stream).
// O snapshot retorna dois momentos (CPUStats e PreCPUStats) para calcular deltas.
func (c *Client) ContainerStats(id string) (*Stats, error) {
	data, err := c.get("/containers/" + id + "/stats?stream=false")
	if err != nil {
		return nil, err
	}
	var out Stats
	return &out, json.Unmarshal(data, &out)
}

// ContainerLogs retorna as últimas `tail` linhas de log do container.
//
// O Docker usa um protocolo de multiplexação no corpo da resposta:
// cada "frame" começa com um header de 8 bytes:
//   - byte[0]: tipo (1=stdout, 2=stderr)
//   - byte[1-3]: zeros (reservados)
//   - byte[4-7]: tamanho do payload em big-endian
//
// O código abaixo lê esses frames e extrai apenas o texto (payload),
// descartando os headers de controle.
func (c *Client) ContainerLogs(id string, tail int) (string, error) {
	// stdout=true e stderr=true → captura ambas as saídas
	// tail=N → últimas N linhas (Docker aceita "all" também, mas aqui usamos número)
	path := fmt.Sprintf("/containers/%s/logs?stdout=true&stderr=true&tail=%d", id, tail)
	data, err := c.get(path)
	if err != nil {
		return "", err
	}

	// percorre os frames do stream multiplexado e extrai o texto
	var sb strings.Builder
	b := data
	for len(b) > 8 { // cada frame tem no mínimo 8 bytes de header
		// lê o tamanho do payload dos bytes 4-7 (big-endian de 32 bits)
		size := int(b[4])<<24 | int(b[5])<<16 | int(b[6])<<8 | int(b[7])
		if 8+size > len(b) {
			break // frame incompleto — para de processar
		}
		sb.Write(b[8 : 8+size]) // escreve o payload (texto real do log)
		b = b[8+size:]          // avança para o próximo frame
	}
	return sb.String(), nil
}


// ══════════════════════════════════════════════════════════════════════════════
// FUNÇÕES DE PACOTE — operações de ciclo de vida (atalhos via DefaultClient)
// ══════════════════════════════════════════════════════════════════════════════

// CreateContainer cria um container sem iniciá-lo.
// Preencha opts.Image (obrigatório) e opts.Name (opcional).
// Use StartContainer(id) logo em seguida para colocá-lo em execução.
func CreateContainer(opts ContainerCreateOptions) (*ContainerCreateResponse, error) {
	return DefaultClient.CreateContainer(opts)
}

// RemoveContainer remove um container pelo ID ou nome.
// force=true força a remoção mesmo que o container esteja rodando.
// removeVolumes=true apaga também os volumes anônimos criados junto com o container.
func RemoveContainer(id string, force, removeVolumes bool) error {
	return DefaultClient.RemoveContainer(id, force, removeVolumes)
}

// StartContainer inicia um container que está parado.
// Não retorna erro se o container já estiver em execução (Docker responde 304).
func StartContainer(id string) error {
	return DefaultClient.StartContainer(id)
}

// StopContainer para o container: envia SIGTERM e aguarda até timeout segundos.
// Se timeout ≤ 0, usa o padrão do Docker (10 segundos).
// Após o timeout sem resposta, o Docker envia SIGKILL.
// Não retorna erro se o container já estiver parado (Docker responde 304).
func StopContainer(id string, timeout int) error {
	return DefaultClient.StopContainer(id, timeout)
}

// RestartContainer reinicia o container.
// O Docker para o container (com o mesmo comportamento de StopContainer)
// e depois o inicia novamente.
// Se timeout ≤ 0, usa o padrão do Docker (10 segundos) antes do SIGKILL.
func RestartContainer(id string, timeout int) error {
	return DefaultClient.RestartContainer(id, timeout)
}

// ══════════════════════════════════════════════════════════════════════════════
// MÉTODOS NO CLIENT — implementações das operações de ciclo de vida
// ══════════════════════════════════════════════════════════════════════════════

// CreateContainer cria um container chamando POST /containers/create.
// O nome do container vai como query param (?name=) pois a Docker API exige isso.
// O restante das opções vai serializado como JSON no corpo da requisição.
func (c *Client) CreateContainer(opts ContainerCreateOptions) (*ContainerCreateResponse, error) {
	path := "/containers/create"
	if opts.Name != "" {
		// nome vai na URL, não no JSON — o Docker exige isso por design
		path += "?name=" + opts.Name
	}
	data, err := c.post(path, opts)
	if err != nil {
		return nil, err
	}
	var out ContainerCreateResponse
	return &out, json.Unmarshal(data, &out)
}

// RemoveContainer remove um container chamando DELETE /containers/{id}.
// force=true → Docker envia SIGKILL antes de remover (mesmo se rodando).
// removeVolumes=true → apaga volumes anônimos montados no container (?v=true).
func (c *Client) RemoveContainer(id string, force, removeVolumes bool) error {
	path := fmt.Sprintf("/containers/%s?force=%v&v=%v", id, force, removeVolumes)
	_, err := c.delete(path)
	return err
}

// StartContainer inicia um container chamando POST /containers/{id}/start.
// Docker retorna 204 ao iniciar com sucesso ou 304 se já estava rodando.
// Ambos são tratados como sucesso (sem erro).
func (c *Client) StartContainer(id string) error {
	_, err := c.post("/containers/"+id+"/start", nil)
	return err
}

// StopContainer para o container chamando POST /containers/{id}/stop.
// Se timeout > 0, envia ?t=N para o Docker aguardar N segundos antes do SIGKILL.
// Docker retorna 204 ao parar ou 304 se já estava parado — ambos sem erro.
func (c *Client) StopContainer(id string, timeout int) error {
	path := "/containers/" + id + "/stop"
	if timeout > 0 {
		path += fmt.Sprintf("?t=%d", timeout)
	}
	_, err := c.post(path, nil)
	return err
}

// RestartContainer reinicia o container chamando POST /containers/{id}/restart.
// Internamente o Docker faz um stop (com timeout) seguido de um start.
// Se timeout > 0, aguarda N segundos antes do SIGKILL na fase de stop.
func (c *Client) RestartContainer(id string, timeout int) error {
	path := "/containers/" + id + "/restart"
	if timeout > 0 {
		path += fmt.Sprintf("?t=%d", timeout)
	}
	_, err := c.post(path, nil)
	return err
}

