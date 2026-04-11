# Docker Management API

API REST para gerenciamento de recursos Docker (containers, imagens, volumes, redes e sistema).

## Arquitetura

O projeto segue **Clean Architecture** com as seguintes camadas:

```
internal/
├── domain/           # Entidades e interfaces de domínio
│   ├── entity/       # Container, Image, Volume, Network, System
│   └── client/       # Interfaces dos clientes Docker
├── application/      # Casos de uso / Serviços
│   └── service/      # ContainerService, ImageService, etc.
├── infrastructure/   # Implementações externas
│   └── docker/       # Cliente Docker SDK
└── interfaces/       # Adaptadores de entrada
    └── http/         # Handlers, DTOs, Router (Huma + Chi)
```

## Tecnologias

- **Go 1.24.1**
- **Huma v2** - Framework HTTP com OpenAPI automático
- **Chi v5** - Router HTTP
- **Docker SDK v26.1.5** - Cliente Docker

## Requisitos

- Go 1.24+
- Docker instalado e rodando
- Permissão para acessar o Docker socket

## Configuração do Docker

### Adicionar usuário ao grupo docker

```bash
sudo groupadd docker
sudo usermod -aG docker $USER
```

**Importante:** Faça logout e login novamente para aplicar as permissões.

Para aplicar imediatamente na sessão atual:

```bash
newgrp docker
```

### Verificar permissões

```bash
docker ps
```

## Executando

```bash
# Build
go build ./...

# Executar
go run cmd/api/main.go

# Ou executar main.go na raiz
go run main.go
```

## API Endpoints

A API estará disponível em `http://localhost:9001`

### Documentação OpenAPI

- Swagger UI: `http://localhost:9001/docs`

### Containers

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | /containers | Listar containers |
| GET | /containers/{id} | Inspecionar container |
| POST | /containers | Criar container |
| POST | /containers/{id}/start | Iniciar container |
| POST | /containers/{id}/stop | Parar container |
| POST | /containers/{id}/restart | Reiniciar container |
| POST | /containers/{id}/pause | Pausar container |
| POST | /containers/{id}/unpause | Despausar container |
| POST | /containers/{id}/kill | Matar container |
| DELETE | /containers/{id} | Remover container |
| POST | /containers/{id}/rename | Renomear container |
| GET | /containers/{id}/logs | Obter logs |
| GET | /containers/{id}/stats | Obter estatísticas |
| POST | /containers/{id}/exec | Executar comando |
| GET | /containers/{id}/top | Listar processos |
| POST | /containers/prune | Limpar containers parados |

### Imagens

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | /images | Listar imagens |
| GET | /images/{id} | Inspecionar imagem |
| POST | /images/pull | Baixar imagem |
| DELETE | /images/{id} | Remover imagem |
| POST | /images/{id}/tag | Taguear imagem |
| GET | /images/{id}/history | Histórico da imagem |
| GET | /images/search | Buscar no registry |
| POST | /images/prune | Limpar imagens não usadas |

### Volumes

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | /volumes | Listar volumes |
| GET | /volumes/{name} | Inspecionar volume |
| POST | /volumes | Criar volume |
| DELETE | /volumes/{name} | Remover volume |
| POST | /volumes/prune | Limpar volumes não usados |

### Redes

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | /networks | Listar redes |
| GET | /networks/{id} | Inspecionar rede |
| POST | /networks | Criar rede |
| DELETE | /networks/{id} | Remover rede |
| POST | /networks/{id}/connect | Conectar container |
| POST | /networks/{id}/disconnect | Desconectar container |
| POST | /networks/prune | Limpar redes não usadas |

### Sistema

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | /system/info | Informações do sistema |
| GET | /system/version | Versão do Docker |
| GET | /system/df | Uso de disco |
| POST | /system/prune | Limpar tudo não usado |
| GET | /system/ping | Health check |

## Exemplos

### Listar containers

```bash
curl http://localhost:9001/containers?all=true
```

### Criar e iniciar um container

```bash
# Criar
curl -X POST http://localhost:9001/containers \
  -H "Content-Type: application/json" \
  -d '{"name": "meu-nginx", "image": "nginx:latest"}'

# Iniciar
curl -X POST http://localhost:9001/containers/meu-nginx/start
```

### Baixar uma imagem

```bash
curl -X POST http://localhost:9001/images/pull \
  -H "Content-Type: application/json" \
  -d '{"image": "alpine", "tag": "latest"}'
```

### Executar comando em container

```bash
curl -X POST http://localhost:9001/containers/meu-nginx/exec \
  -H "Content-Type: application/json" \
  -d '{"cmd": ["ls", "-la", "/"]}'
```

## Desenvolvimento

### Debug

Configure `GOTOOLCHAIN=local` no VS Code para evitar problemas de versão:

```json
{
  "go.toolsEnvVars": {
    "GOTOOLCHAIN": "local"
  }
}
```

### Estrutura de arquivos

```
.
├── cmd/
│   └── api/
│       └── main.go          # Entrada alternativa
├── internal/
│   ├── domain/
│   │   ├── entity/          # Entidades de domínio
│   │   └── client/          # Interfaces
│   ├── application/
│   │   └── service/         # Lógica de negócio
│   ├── infrastructure/
│   │   └── docker/          # Implementação Docker SDK
│   └── interfaces/
│       └── http/
│           ├── dto/         # Data Transfer Objects
│           ├── handler/     # HTTP Handlers
│           └── router.go    # Rotas
├── main.go                  # Entrada principal
├── go.mod
├── go.sum
└── README.md
```

## Licença

MIT
