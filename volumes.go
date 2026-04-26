package godocker

import "encoding/json" // json.Unmarshal — converte JSON bytes em structs Go

// ══════════════════════════════════════════════════════════════════════════════
// FUNÇÕES DE PACOTE (atalho via DefaultClient)
// ══════════════════════════════════════════════════════════════════════════════

// ListVolumes lista todos os volumes Docker criados na máquina.
// Equivale a executar `docker volume ls` no terminal.
// Volumes são usados para persistir dados além do ciclo de vida dos containers.
func ListVolumes() ([]Volume, error) {
	return DefaultClient.ListVolumes()
}

// ══════════════════════════════════════════════════════════════════════════════
// MÉTODO NO CLIENT
// ══════════════════════════════════════════════════════════════════════════════

// ListVolumes lista os volumes chamando GET /volumes.
//
// O endpoint retorna um objeto com a chave "Volumes" contendo o array,
// não um array direto como os outros endpoints. Por isso usamos uma
// struct anônima intermediária para "desembrulhar" o JSON:
//
//	{ "Volumes": [...], "Warnings": [...] }
//	                ↓ Unmarshal
//	result.Volumes → []Volume
func (c *Client) ListVolumes() ([]Volume, error) {
	data, err := c.get("/volumes")
	if err != nil {
		return nil, err
	}
	// struct intermediária que espelha o formato da resposta do Docker
	var result struct {
		Volumes []Volume `json:"Volumes"` // array de volumes
		// Warnings são ignorados aqui — use VolumeListResponse se precisar deles
	}
	return result.Volumes, json.Unmarshal(data, &result)
}
