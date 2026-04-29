package godocker

import "encoding/json" // json.Unmarshal — converte JSON bytes em structs Go

// ══════════════════════════════════════════════════════════════════════════════
// FUNÇÕES DE PACOTE (atalho via DefaultClient)
// ══════════════════════════════════════════════════════════════════════════════

// ListImages lista todas as imagens Docker armazenadas localmente na máquina.
// Equivale a executar `docker images` no terminal.
// Retorna informações como ID, tags, tamanho e quantos containers usam cada imagem.
func ListImages() ([]Image, error) {
	return DefaultClient.ListImages()
}

// ══════════════════════════════════════════════════════════════════════════════
// MÉTODO NO CLIENT
// ══════════════════════════════════════════════════════════════════════════════

// ListImages lista as imagens locais chamando GET /images/json.
// O Docker retorna um array JSON com todas as imagens disponíveis.
func (c *Client) ListImages() ([]Image, error) {
	data, err := c.get("/images/json")
	if err != nil {
		return nil, err
	}
	// deserializa o array JSON em um slice de Image
	var out []Image
	return out, json.Unmarshal(data, &out)
}


