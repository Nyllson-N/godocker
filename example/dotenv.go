package main

import (
	"bufio"  // bufio.NewScanner — lê o arquivo linha por linha sem carregar tudo na memória
	"os"     // os.Open, os.Setenv, os.IsNotExist — acesso ao sistema de arquivos e ao ambiente
	"strings" // strings.TrimSpace, strings.HasPrefix, strings.Cut — manipulação de strings
)

// loadDotEnv lê um arquivo .env e carrega as variáveis como variáveis de ambiente do processo.
//
// Regras do formato suportado:
//   - Cada linha no formato KEY=valor define uma variável
//   - Linhas começando com # são comentários e são ignoradas
//   - Linhas vazias são ignoradas
//   - Valores podem ter aspas duplas: KEY="valor com espaços" → valor com espaços
//   - Variáveis já definidas no ambiente do sistema NÃO são sobrescritas
//     (isso permite sobrescrever via shell: PORT=9090 go run . tem prioridade sobre o .env)
//
// O arquivo é opcional — se não existir, retorna nil sem erro.
func loadDotEnv(path string) error {
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

		// ignora linhas vazias e comentários
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// divide na primeira ocorrência de "=" para separar chave e valor
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue // linha sem "=" — formato inválido, ignora
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		// remove aspas duplas ao redor do valor: "meu valor" → meu valor
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}

		// não sobrescreve variável já definida no ambiente do sistema operacional
		// isso garante que variáveis de shell têm prioridade sobre o .env
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return scanner.Err()
}
