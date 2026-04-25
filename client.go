// Package godocker fornece um cliente para a Docker Engine API
// sem dependências externas, com detecção automática de ambiente
// (Linux nativo, WSL2 e Windows/Docker Desktop).
package godocker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// ══════════════════════════════════════════════════════════════════════════════
// TRANSPORT
// ══════════════════════════════════════════════════════════════════════════════

// APIVersion padrão. Pode ser sobrescrito via DOCKER_API_VERSION ou Client.APIVersion.
const defaultAPIVersion = "v1.47"

// Client é o cliente Docker. Use New() ou DefaultClient.
type Client struct {
	mode       string // "unix" | "tcp"
	baseURL    string
	APIVersion string // ex: "v1.44", "v1.47" — padrão: defaultAPIVersion
	http       *http.Client
}

// Mode retorna o modo de conexão detectado: "unix" ou "tcp".
func (c *Client) Mode() string { return c.mode }

// BaseURL retorna a URL base usada nas requisições.
func (c *Client) BaseURL() string { return c.baseURL }

// apiBase retorna o prefixo versionado: /v1.47
func (c *Client) apiBase() string {
	if c.APIVersion != "" {
		return "/" + c.APIVersion
	}
	if v := os.Getenv("DOCKER_API_VERSION"); v != "" {
		return "/" + v
	}
	return "/" + defaultAPIVersion
}

// DefaultClient é o cliente global inicializado automaticamente na startup.
// Detecta o ambiente uma única vez: DOCKER_HOST → Windows → WSL2 → Linux → fallback TCP.
var DefaultClient = New()

// New cria um novo Client com detecção automática de ambiente.
// A ordem de prioridade é:
//  1. Variável de ambiente DOCKER_HOST
//  2. Windows nativo      → TCP localhost:2375
//  3. WSL2                → TCP localhost:2375 (se alcançável)
//  4. Linux nativo        → Unix socket /var/run/docker.sock
//  5. Fallback            → TCP localhost:2375
func New() *Client {
	// 1. DOCKER_HOST override
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		raw := host
		if strings.HasPrefix(raw, "unix://") {
			return newUnix(strings.TrimPrefix(raw, "unix://"))
		}
		raw = strings.TrimPrefix(raw, "tcp://")
		raw = strings.TrimPrefix(raw, "http://")
		return newTCP("http://" + raw)
	}

	// 2. Windows nativo
	if runtime.GOOS == "windows" {
		return newTCP("http://localhost:2375")
	}

	// 3. WSL2
	if isWSL() && canReachTCP("localhost:2375") {
		return newTCP("http://localhost:2375")
	}

	// 4. Linux nativo — Unix socket
	const sockPath = "/var/run/docker.sock"
	if _, err := os.Stat(sockPath); err == nil {
		return newUnix(sockPath)
	}

	// 5. Fallback
	return newTCP("http://localhost:2375")
}

// NewTCP cria um Client apontando para um endereço TCP específico.
// Ex: godocker.NewTCP("localhost:2375")
func NewTCP(addr string) *Client {
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	return newTCP(addr)
}

// NewUnix cria um Client usando um socket Unix específico.
// Ex: godocker.NewUnix("/var/run/docker.sock")
func NewUnix(sockPath string) *Client {
	return newUnix(sockPath)
}

func newUnix(sockPath string) *Client {
	return &Client{
		mode:    "unix",
		baseURL: "http://docker",
		http: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{}).DialContext(ctx, "unix", sockPath)
				},
			},
		},
	}
}

func newTCP(baseURL string) *Client {
	return &Client{
		mode:    "tcp",
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

func canReachTCP(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ══════════════════════════════════════════════════════════════════════════════
// HTTP HELPERS (internos)
// ══════════════════════════════════════════════════════════════════════════════

func (c *Client) request(method, path string, body any) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("docker [%s] %s %s: %w", c.mode, method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, fmt.Errorf("docker API %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return data, resp.StatusCode, nil
}

func (c *Client) get(path string) ([]byte, error) {
	data, _, err := c.request("GET", c.apiBase()+path, nil)
	return data, err
}

func (c *Client) post(path string, body any) ([]byte, error) {
	data, _, err := c.request("POST", c.apiBase()+path, body)
	return data, err
}

func (c *Client) delete(path string) (int, error) {
	_, code, err := c.request("DELETE", c.apiBase()+path, nil)
	return code, err
}