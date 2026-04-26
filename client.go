// Package godocker é uma biblioteca Go para se comunicar com o Docker Engine API.
// Não tem dependências externas — usa apenas a biblioteca padrão do Go.
// Detecta automaticamente o ambiente: Linux, WSL2 ou Windows/Docker Desktop.
package godocker

import (
	"bytes"         // bytes.NewReader — transforma []byte em um io.Reader para o corpo do request
	"context"       // context.Context — usado no DialContext para cancelamento de conexão
	"encoding/json" // json.Marshal / json.Unmarshal — serializa Go → JSON e JSON → Go
	"fmt"           // fmt.Errorf — formata mensagens de erro
	"io"            // io.Reader, io.ReadAll — leitura do corpo da resposta HTTP
	"net"           // net.Conn, net.Dialer, net.DialTimeout — conexões TCP e Unix socket
	"net/http"      // http.Client, http.Request, http.Transport — cliente HTTP padrão
	"os"            // os.Getenv, os.Stat, os.ReadFile — leitura de variáveis de ambiente e arquivos
	"runtime"       // runtime.GOOS — detecta o sistema operacional em tempo de execução
	"strconv"       // strconv.Atoi — converte string para inteiro (usado na comparação de versões)
	"strings"       // strings.HasPrefix, strings.TrimPrefix, strings.Contains — manipulação de strings
	"time"          // time.Second, time.Duration — timeouts de conexão
)

// ══════════════════════════════════════════════════════════════════════════════
// CONSTANTES DE VERSÃO
// ══════════════════════════════════════════════════════════════════════════════

// defaultAPIVersion é a versão da Docker Engine API usada por padrão.
// Toda requisição vai para uma URL como /v1.47/containers/json.
// Pode ser sobrescrito pela variável de ambiente DOCKER_API_VERSION
// ou pelo campo APIVersion do Client.
const defaultAPIVersion = "v1.47"

// minAPIVersion é a versão mínima aceita pelo daemon Docker moderno.
// Se o usuário configurar uma versão mais antiga (ex: 1.41 via variável de ambiente),
// ela será automaticamente elevada para esta versão para evitar o erro 400 do Docker.
const minAPIVersion = "v1.44"

// ══════════════════════════════════════════════════════════════════════════════
// TIPO CLIENT
// ══════════════════════════════════════════════════════════════════════════════

// Client representa uma conexão com o daemon Docker.
// Guarda o modo de transporte (unix socket ou TCP), a URL base
// e o cliente HTTP interno que faz as requisições.
type Client struct {
	mode       string      // "unix" quando usa socket, "tcp" quando usa rede TCP
	baseURL    string      // URL base — ex: "http://docker" (unix) ou "http://localhost:2375" (tcp)
	APIVersion string      // versão da API a usar — ex: "v1.44". Vazio = usa defaultAPIVersion
	http       *http.Client // cliente HTTP da biblioteca padrão, com timeout configurado
}

// Mode retorna o tipo de conexão que está sendo usada: "unix" ou "tcp".
// Útil para saber se o cliente está usando socket local ou TCP remoto.
func (c *Client) Mode() string { return c.mode }

// BaseURL retorna o endereço base das requisições.
// Ex: "http://docker" para socket Unix ou "http://localhost:2375" para TCP.
func (c *Client) BaseURL() string { return c.baseURL }

// apiBase monta o prefixo de versão que vai em cada URL de requisição.
// Exemplo de retorno: "/v1.47"
// Prioridade de escolha da versão:
//  1. Campo APIVersion do Client (configurado manualmente)
//  2. Variável de ambiente DOCKER_API_VERSION
//  3. defaultAPIVersion (v1.47)
//
// Em todos os casos, clampAPIVersion garante que a versão não seja menor que v1.44.
func (c *Client) apiBase() string {
	if c.APIVersion != "" {
		// o usuário configurou a versão manualmente no struct
		return "/" + clampAPIVersion(c.APIVersion)
	}
	if v := os.Getenv("DOCKER_API_VERSION"); v != "" {
		// variável de ambiente definida — ex: DOCKER_API_VERSION=1.41
		return "/" + clampAPIVersion(v)
	}
	// nenhuma configuração → usa o padrão
	return "/" + defaultAPIVersion
}

// clampAPIVersion garante que a versão nunca seja inferior a minAPIVersion.
// Também normaliza o prefixo "v": "1.41" → "v1.41" → corrigido para "v1.44".
func clampAPIVersion(v string) string {
	// adiciona o "v" na frente se estiver faltando (ex: "1.44" → "v1.44")
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	// se a versão for mais antiga que o mínimo, eleva para o mínimo
	if apiVersionLess(v, minAPIVersion) {
		return minAPIVersion
	}
	return v
}

// apiVersionLess compara duas versões Docker no formato "vMAJOR.MINOR".
// Retorna true se a versão 'a' for mais antiga que a versão 'b'.
// Exemplo: apiVersionLess("v1.41", "v1.44") → true
func apiVersionLess(a, b string) bool {
	// parse extrai os números major e minor de uma string como "v1.44"
	parse := func(s string) (int, int) {
		s = strings.TrimPrefix(s, "v")      // remove o "v" inicial: "v1.44" → "1.44"
		parts := strings.SplitN(s, ".", 2)  // divide em ["1", "44"]
		if len(parts) < 2 {
			return 0, 0
		}
		maj, _ := strconv.Atoi(parts[0]) // converte "1" → 1
		min, _ := strconv.Atoi(parts[1]) // converte "44" → 44
		return maj, min
	}
	majA, minA := parse(a)
	majB, minB := parse(b)
	if majA != majB {
		return majA < majB // compara major primeiro (ex: v1 vs v2)
	}
	return minA < minB // se major igual, compara minor (ex: v1.41 vs v1.44)
}

// ══════════════════════════════════════════════════════════════════════════════
// CLIENTE PADRÃO E CONSTRUTORES
// ══════════════════════════════════════════════════════════════════════════════

// DefaultClient é o cliente global criado automaticamente quando o pacote é importado.
// Detecta o ambiente uma única vez na inicialização do programa.
// Na maioria dos casos é tudo que você precisa — basta chamar docker.Info(), docker.ListContainers(), etc.
var DefaultClient = New()

// New cria um novo Client com detecção automática de ambiente.
// A detecção segue esta ordem de prioridade:
//
//  1. DOCKER_HOST (variável de ambiente) — ex: DOCKER_HOST=tcp://192.168.1.5:2375
//  2. Windows nativo → TCP localhost:2375 (Docker Desktop expõe esta porta)
//  3. WSL2 + Docker Desktop ativo → TCP localhost:2375 (acessível do WSL)
//  4. Linux com socket disponível → Unix /var/run/docker.sock (mais eficiente)
//  5. Fallback → TCP localhost:2375
func New() *Client {
	// 1. DOCKER_HOST tem a maior prioridade — sobrescreve tudo
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		raw := host
		// socket Unix: unix:///var/run/docker.sock → extrai só o caminho
		if path, ok := strings.CutPrefix(raw, "unix://"); ok {
			return newUnix(path)
		}
		// TCP: remove prefixos "tcp://" ou "http://" e reconstrói com "http://"
		raw = strings.TrimPrefix(raw, "tcp://")
		raw = strings.TrimPrefix(raw, "http://")
		return newTCP("http://" + raw)
	}

	// 2. Windows nativo: o runtime.GOOS retorna "windows"
	// Docker Desktop no Windows expõe o daemon em TCP sem precisar de socket
	if runtime.GOOS == "windows" {
		return newTCP("http://localhost:2375")
	}

	// 3. WSL2: detecta se está rodando dentro do subsistema Linux do Windows
	// e verifica se o Docker Desktop está acessível via TCP
	if isWSL() && canReachTCP("localhost:2375") {
		return newTCP("http://localhost:2375")
	}

	// 4. Linux nativo: verifica se o socket Unix existe no caminho padrão
	// O socket é a forma mais eficiente de comunicação — sem overhead de rede
	const sockPath = "/var/run/docker.sock"
	if _, err := os.Stat(sockPath); err == nil {
		return newUnix(sockPath)
	}

	// 5. Fallback: tenta TCP, mesmo sem certeza que está disponível
	return newTCP("http://localhost:2375")
}

// NewTCP cria um Client que se conecta via TCP em um endereço específico.
// Útil para conectar em Docker remoto ou Docker Desktop com TCP exposto.
// Exemplo: docker.NewTCP("192.168.1.100:2375")
func NewTCP(addr string) *Client {
	// garante que a URL começa com "http://" para o http.Client funcionar
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	return newTCP(addr)
}

// NewUnix cria um Client que se conecta via socket Unix em um caminho específico.
// Útil quando o socket não está no caminho padrão (/var/run/docker.sock).
// Exemplo: docker.NewUnix("/run/user/1000/docker.sock")
func NewUnix(sockPath string) *Client {
	return newUnix(sockPath)
}

// newUnix é o construtor interno para conexões Unix socket.
// O truque aqui é usar um http.Transport customizado que redireciona
// qualquer conexão HTTP para o socket Unix em vez da rede TCP.
// A URL "http://docker" é fictícia — o Docker nunca vê essa URL,
// o DialContext intercepta antes e abre o socket.
func newUnix(sockPath string) *Client {
	return &Client{
		mode:    "unix",
		baseURL: "http://docker", // URL fictícia necessária para o http.Client
		http: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				// DialContext substitui a abertura normal de conexão TCP.
				// Quando o http.Client tenta conectar, chamamos este função
				// que abre o socket Unix em vez de uma conexão de rede.
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "unix", sockPath)
				},
			},
		},
	}
}

// newTCP é o construtor interno para conexões TCP.
// Usa o http.Client padrão do Go, apenas com timeout configurado.
func newTCP(baseURL string) *Client {
	return &Client{
		mode:    "tcp",
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// ══════════════════════════════════════════════════════════════════════════════
// FUNÇÕES DE DETECÇÃO DE AMBIENTE
// ══════════════════════════════════════════════════════════════════════════════

// isWSL verifica se o código está sendo executado dentro do WSL2 (Windows Subsystem for Linux).
// O WSL2 escreve "microsoft" ou "wsl" no arquivo /proc/version do kernel Linux.
func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false // arquivo não existe → não é Linux (nem WSL)
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// canReachTCP verifica se é possível abrir uma conexão TCP no endereço dado.
// Usado para testar se o Docker Desktop está escutando antes de tentar usá-lo.
// Timeout de 2 segundos para não travar o startup da aplicação.
func canReachTCP(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close() // fecha imediatamente — só precisávamos saber se conectou
	return true
}

// ══════════════════════════════════════════════════════════════════════════════
// HELPERS HTTP INTERNOS
// ══════════════════════════════════════════════════════════════════════════════

// request é o método central que executa todas as chamadas HTTP para o Docker.
// Todos os outros métodos (get, post, delete) chamam este.
//
// Parâmetros:
//   - method: "GET", "POST", "DELETE"
//   - path:   URL completa incluindo versão — ex: "/v1.47/containers/json"
//   - body:   estrutura Go que será serializada como JSON no corpo (pode ser nil)
//
// Retorna:
//   - []byte:  corpo da resposta do Docker (JSON bruto)
//   - int:     código HTTP da resposta (200, 404, 500, etc.)
//   - error:   erro de rede ou erro de API (status >= 400)
func (c *Client) request(method, path string, body any) ([]byte, int, error) {
	// serializa o body para JSON se houver (usado nos POSTs)
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body) // Go struct → JSON bytes
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b) // transforma []byte em io.Reader
	}

	// monta a requisição HTTP: método + URL completa + corpo
	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	// informa ao Docker que estamos enviando JSON (necessário nos POSTs)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// executa a requisição — o http.Client faz a conexão (unix ou tcp)
	resp, err := c.http.Do(req)
	if err != nil {
		// erro de rede: sem conexão, timeout, socket não existe, etc.
		return nil, 0, fmt.Errorf("docker [%s] %s %s: %w", c.mode, method, path, err)
	}
	defer resp.Body.Close() // garante que o corpo é fechado mesmo em caso de erro

	// lê todo o corpo da resposta de uma vez
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	// status >= 400 significa erro da API Docker (ex: container não encontrado, versão antiga)
	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, fmt.Errorf("docker API %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return data, resp.StatusCode, nil
}

// get executa um GET no endpoint Docker adicionando o prefixo de versão.
// Exemplo: c.get("/containers/json") → GET /v1.47/containers/json
func (c *Client) get(path string) ([]byte, error) {
	data, _, err := c.request("GET", c.apiBase()+path, nil)
	return data, err
}

// post executa um POST no endpoint Docker com um corpo JSON.
// Exemplo: c.post("/networks/create", opts) → POST /v1.47/networks/create
func (c *Client) post(path string, body any) ([]byte, error) {
	data, _, err := c.request("POST", c.apiBase()+path, body)
	return data, err
}

// delete executa um DELETE no endpoint Docker.
// Retorna o código HTTP pois alguns DELETEs retornam 204 (sem corpo).
func (c *Client) delete(path string) (int, error) {
	_, code, err := c.request("DELETE", c.apiBase()+path, nil)
	return code, err
}

// ══════════════════════════════════════════════════════════════════════════════
// API BRUTA
// ══════════════════════════════════════════════════════════════════════════════

// RawGet faz um GET no endpoint Docker e retorna o JSON exatamente como
// o daemon enviou, sem passar por nenhuma struct de tipagem.
// Útil para acessar campos que não estão mapeados nos tipos ou para
// chamar endpoints não cobertos pela biblioteca.
//
// Exemplos:
//
//	client.RawGet("/info")
//	client.RawGet("/version")
//	client.RawGet("/containers/abc123/json")
func (c *Client) RawGet(path string) (json.RawMessage, error) {
	return c.get(path)
}

// RawGet faz um GET no endpoint usando o DefaultClient.
// Permite chamar docker.RawGet("/info") sem criar um Client manualmente.
func RawGet(path string) (json.RawMessage, error) {
	return DefaultClient.RawGet(path)
}
