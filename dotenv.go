package godocker

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv lê um arquivo .env e carrega as variáveis como variáveis de ambiente do processo.
// Chame esta função no início do seu main() antes de usar qualquer função da biblioteca.
//
// O arquivo é lido de onde o binário está sendo executado (diretório do projeto importador),
// não de onde a biblioteca está instalada.
//
// Variáveis já definidas no ambiente do sistema operacional (via shell) NÃO são sobrescritas —
// isso permite sobrescrever via shell sem editar o arquivo: DOCKER_HOST=... go run .
//
// Variáveis suportadas pela biblioteca:
//   - DOCKER_HOST       — endereço de uma máquina Docker (ex: tcp://192.168.1.10:2375)
//   - DOCKER_API_VERSION — versão da API Docker (ex: 1.47 — elevada automaticamente se muito antiga)
//
// Formato do arquivo .env:
//
//	KEY=valor            # básico
//	KEY="valor com espaços"  # com aspas duplas
//	# comentário         # ignorado
//
// Uso:
//
//	func main() {
//	    docker.LoadEnv(".env")     // carrega do diretório atual do projeto
//	    info, err := docker.Info()
//	    ...
//	}
func LoadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // .env é opcional — sem arquivo não é erro
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return scanner.Err()
}
