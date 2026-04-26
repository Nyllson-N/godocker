# Rotas da API

Servidor rodando em `http://localhost:8080`.  
Para iniciar: `cd example && go run .`

---

## GET /

Lista todas as rotas disponíveis no servidor.

```bash
curl http://localhost:8080/
```

---

## GET /info

Retorna informações completas do daemon Docker (versão, sistema operacional, número de containers, memória total, etc.).

```bash
curl http://localhost:8080/info
```

---

## GET /containers

Lista containers em execução. Adicione `?all=1` para incluir os parados.

```bash
# apenas containers rodando
curl http://localhost:8080/containers

# todos os containers (incluindo parados)
curl "http://localhost:8080/containers?all=1"
```

---

## GET /containers/{id}

Retorna o inspect completo de um container: configuração, estado, rede, volumes, healthcheck, etc.  
`{id}` pode ser o ID completo, os primeiros caracteres do ID ou o nome do container.

```bash
curl http://localhost:8080/containers/meu-nginx
curl http://localhost:8080/containers/a1b2c3d4e5f6
```

---

## GET /containers/{id}/stats

Retorna um snapshot de uso de recursos do container: CPU, memória, I/O de disco e tráfego de rede.  
Só funciona para containers em execução.

```bash
curl http://localhost:8080/containers/meu-nginx/stats
```

---

## GET /containers/{id}/logs

Retorna as últimas linhas de log do container (stdout + stderr juntos).  
`?tail=N` controla quantas linhas retornar (padrão: 100).

```bash
# últimas 100 linhas (padrão)
curl http://localhost:8080/containers/meu-nginx/logs

# últimas 50 linhas
curl "http://localhost:8080/containers/meu-nginx/logs?tail=50"

# todas as linhas
curl "http://localhost:8080/containers/meu-nginx/logs?tail=0"
```

---

## POST /containers

Cria um container sem iniciá-lo. O corpo deve ser um JSON com as opções.  
Use `?name=` na URL para nomear o container. Em seguida, use `/start` para iniciá-lo.

**Campos do body:**

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `Image` | string | Imagem a usar — **obrigatório** (ex: `"nginx:latest"`) |
| `Cmd` | string[] | Comando a executar (ex: `["nginx", "-g", "daemon off;"]`) |
| `Env` | string[] | Variáveis de ambiente no formato `"KEY=valor"` |
| `WorkingDir` | string | Diretório de trabalho dentro do container |
| `Hostname` | string | Hostname do container |
| `Labels` | object | Labels para organização (ex: `{"app": "web"}`) |
| `ExposedPorts` | object | Portas declaradas (ex: `{"80/tcp": {}}`) |
| `HostConfig` | object | Configurações de runtime: portas, volumes, restart, limites |

**Exemplos:**

```bash
# container nginx simples
curl -X POST "http://localhost:8080/containers?name=meu-nginx" \
  -H "Content-Type: application/json" \
  -d '{"Image": "nginx:latest"}'

# container com porta exposta e variáveis de ambiente
curl -X POST "http://localhost:8080/containers?name=meu-app" \
  -H "Content-Type: application/json" \
  -d '{
    "Image": "nginx:latest",
    "Env": ["NGINX_PORT=80"],
    "ExposedPorts": {"80/tcp": {}},
    "HostConfig": {
      "PortBindings": {
        "80/tcp": [{"HostIp": "0.0.0.0", "HostPort": "8090"}]
      },
      "RestartPolicy": {"Name": "unless-stopped", "MaximumRetryCount": 0}
    }
  }'

# container com volume montado
curl -X POST "http://localhost:8080/containers?name=meu-db" \
  -H "Content-Type: application/json" \
  -d '{
    "Image": "postgres:16",
    "Env": ["POSTGRES_PASSWORD=senha123"],
    "HostConfig": {
      "Binds": ["/meu/dados:/var/lib/postgresql/data"]
    }
  }'
```

**Resposta (HTTP 201):**

```json
{
  "data": {
    "Id": "a1b2c3d4e5f6...",
    "Warnings": []
  },
  "error": null
}
```

---

## DELETE /containers/{id}

Remove um container pelo ID ou nome.

| Parâmetro | Valor | Descrição |
|-----------|-------|-----------|
| `?force=1` | 1 | Remove mesmo que o container esteja rodando |
| `?volumes=1` | 1 | Remove também os volumes anônimos do container |

```bash
# remoção simples (container deve estar parado)
curl -X DELETE http://localhost:8080/containers/meu-nginx

# força remoção mesmo rodando
curl -X DELETE "http://localhost:8080/containers/meu-nginx?force=1"

# remove container e seus volumes anônimos
curl -X DELETE "http://localhost:8080/containers/meu-nginx?force=1&volumes=1"
```

---

## POST /containers/{id}/start

Inicia um container que está parado. Não retorna erro se o container já estiver em execução.

```bash
curl -X POST http://localhost:8080/containers/meu-nginx/start
```

**Fluxo completo — criar e iniciar:**

```bash
# 1. cria o container
curl -X POST "http://localhost:8080/containers?name=meu-nginx" \
  -H "Content-Type: application/json" \
  -d '{"Image": "nginx:latest", "HostConfig": {"PortBindings": {"80/tcp": [{"HostIp": "0.0.0.0", "HostPort": "8090"}]}}}'

# 2. inicia o container criado
curl -X POST http://localhost:8080/containers/meu-nginx/start
```

---

## POST /containers/{id}/stop

Para um container em execução. Envia SIGTERM e aguarda o processo encerrar.  
`?t=N` define quantos segundos aguardar antes de enviar SIGKILL (padrão do Docker: 10s).

```bash
# para com timeout padrão (10 segundos)
curl -X POST http://localhost:8080/containers/meu-nginx/stop

# para aguardando até 30 segundos
curl -X POST "http://localhost:8080/containers/meu-nginx/stop?t=30"

# para imediatamente (SIGKILL direto)
curl -X POST "http://localhost:8080/containers/meu-nginx/stop?t=1"
```

---

## POST /containers/{id}/restart

Reinicia um container (equivale a stop + start). Aceita o mesmo `?t=N` de stop.

```bash
# reinicia com timeout padrão
curl -X POST http://localhost:8080/containers/meu-nginx/restart

# reinicia aguardando até 15 segundos no stop
curl -X POST "http://localhost:8080/containers/meu-nginx/restart?t=15"
```

---

## GET /images

Lista todas as imagens Docker armazenadas localmente.

```bash
curl http://localhost:8080/images
```

---

## GET /networks

Lista todas as redes Docker disponíveis.

```bash
curl http://localhost:8080/networks
```

---

## GET /networks/{id}

Retorna o inspect completo de uma rede: driver, configuração IPAM, containers conectados, etc.

```bash
curl http://localhost:8080/networks/bridge
curl http://localhost:8080/networks/minha-rede
```

---

## GET /volumes

Lista todos os volumes Docker criados na máquina.

```bash
curl http://localhost:8080/volumes
```

---

## GET /all

Retorna todos os dados em uma única requisição: daemon + containers + images + networks + volumes.  
Útil para carregar um dashboard completo com uma só chamada.

```bash
# apenas containers rodando
curl http://localhost:8080/all

# inclui containers parados
curl "http://localhost:8080/all?all=1"
```

---

## Formato das respostas

Todas as rotas retornam o mesmo envelope JSON:

```json
// sucesso
{
  "data": { ... },
  "error": null
}

// erro
{
  "data": null,
  "error": "docker API 404: No such container: meu-nginx"
}
```
