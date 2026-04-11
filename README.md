# FlagFlash

Plataforma multi-tenant de **feature flags** e configuração dinâmica com dashboard web, API REST, WebSocket em tempo real e SDK Go.

## Visão Geral

```
Dashboard (React/TS) + SDK Clients
          ↓
  API (Go, Chi + Huma v2)
  REST  |  WebSocket
  JWT   |  API Key
          ↓
  PostgreSQL  +  Redis (cache + pub/sub)
```

**Hierarquia de domínio:**

```
Tenant → Applications → Environments → Feature Flags (boolean / json / string / number)
                                     ↳ Targeting Rules
                                     ↳ API Keys
```

---

## Início Rápido

### Pré-requisitos

- Docker & Docker Compose
- Go 1.21+
- Node.js 18+

### 1. Subir infraestrutura (PostgreSQL + Redis)

```bash
docker-compose up -d
```

> As migrations são aplicadas automaticamente ao iniciar o container do Postgres.

### 2. Rodar a API

```bash
cd api
export JWT_SECRET="sua-chave-secreta-de-pelo-menos-32-caracteres"
go run ./cmd/api
```

A API sobe em `http://localhost:9001`.

### 3. Rodar o Dashboard

```bash
cd app
npm install
npm run dev
```

O frontend sobe em `http://localhost:5173`.

### Acesso padrão

| URL | Descrição |
|-----|----------|
| `http://localhost:5173` | Dashboard (dev) |
| `http://localhost:9001/api/v1/flagflash` | API |
| `http://localhost:9001/api/v1/flagflash/docs` | OpenAPI (Swagger UI) |
| `http://localhost:9001/health` | Health check |

**Login padrão:** `admin@flagflash.io` / `admin123`

---

## Variáveis de Ambiente (API)

| Variável | Padrão | Obrigatório | Descrição |
|----------|--------|:-----------:|-----------|
| `JWT_SECRET` | — | ✅ | Segredo JWT (app falha ao iniciar se ausente) |
| `JWT_EXPIRATION` | `24h` | — | Validade do token |
| `SERVER_PORT` | `9001` | — | Porta HTTP |
| `APP_ENV` | `development` | — | `development` ou `production` |
| `CORS_ALLOWED_ORIGINS` | `*` | — | Origens permitidas (vírgula-separadas) |
| `DB_HOST` | `localhost` | — | Host do PostgreSQL |
| `DB_PORT` | `5432` | — | Porta do PostgreSQL |
| `DB_USER` | `flagflash` | — | Usuário do banco |
| `DB_PASSWORD` | `flagflash` | — | Senha do banco |
| `DB_NAME` | `flagflash` | — | Nome do banco |
| `DB_SSL_MODE` | `prefer` | — | Modo SSL do Postgres |
| `REDIS_HOST` | `localhost` | — | Host do Redis |
| `REDIS_PORT` | `6379` | — | Porta do Redis |
| `REDIS_PASSWORD` | `""` | — | Senha do Redis |

> As variáveis podem ser definidas em um arquivo `.env` na raiz ou na pasta `api/`.

---

## Estrutura do Projeto

```
flagflash/
├── docker-compose.yml       # PostgreSQL + Redis
├── api/                     # Backend Go (Clean Architecture)
│   ├── cmd/api/main.go      # Ponto de entrada
│   ├── internal/
│   │   ├── domain/          # Entidades e interfaces de repositório
│   │   ├── application/     # Serviços de negócio
│   │   ├── infrastructure/  # Postgres, Redis, WebSocket
│   │   └── interfaces/http/ # Handlers HTTP, middlewares, roteador
│   ├── migrations/          # SQL migrations
│   └── pkg/                 # auth, logger, middleware
├── app/                     # Frontend React + TypeScript + Tailwind CSS
│   └── src/
│       ├── pages/flagflash/ # Páginas do dashboard
│       ├── services/        # Clientes HTTP (flagflash-api.ts)
│       ├── contexts/        # AuthContext
│       └── types/           # Tipos TypeScript
├── sdk/                     # Go SDK para clientes da plataforma
│   └── example/basic/       # Exemplo de uso do SDK
└── docs/
    └── ARCHITECTURE.md      # Documento de arquitetura detalhado
```

---

## Rotas da API

Todas as rotas são prefixadas com `/api/v1/flagflash`.

### Autenticação (público)

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/auth/login` | Login com email/senha → JWT |
| `POST` | `/auth/register` | Registrar tenant + usuário owner |
| `POST` | `/auth/refresh` | Renovar token |
| `POST` | `/auth/switch-tenant` | Trocar tenant ativo |
| `POST` | `/auth/change-password` | Alterar senha |

### SDK (autenticação por API Key)

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/sdk/flags` | Listar todos os flags do ambiente |
| `GET` | `/sdk/flags/{key}` | Buscar flag por chave |
| `POST` | `/sdk/evaluate` | Avaliar flag com contexto de targeting |
| `POST` | `/sdk/evaluate-all` | Avaliar todos os flags |
| `GET` | `/sdk/ws` | WebSocket (atualizações em tempo real) |

### Dashboard `/manage` (JWT obrigatório)

| Recurso | Rotas |
|---------|-------|
| Tenants | CRUD + listar membros |
| Applications | CRUD por tenant |
| Environments | CRUD por application |
| Feature Flags | CRUD, toggle, copiar entre ambientes |
| Targeting Rules | CRUD por flag |
| API Keys | Gerar, revogar, rotacionar |
| Users | Gerenciamento de usuários e papéis |
| Audit Log | Histórico de mudanças |
| Usage Metrics | Estatísticas de avaliação |

---

## SDK Go

```bash
go get github.com/IzuCas/flagflash/sdk
```

### Exemplo básico

```go
import (
    "context"
    "time"
    sdk "github.com/IzuCas/flagflash/sdk"
)

func main() {
    client := sdk.New(
        "ff_sua_api_key",
        "http://localhost:9001",
        sdk.WithTimeout(3*time.Second),
    )

    ctx := context.Background()

    // Conecta e preenche cache local via WebSocket
    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    defer client.Close()

    // Verificação booleana (leitura do cache — latência zero)
    if client.IsEnabled(ctx, "dark_mode") {
        fmt.Println("dark mode ativo")
    }

    // Avaliação com contexto de targeting (enviado ao servidor)
    result, _ := client.Evaluate(ctx, "novo-checkout", sdk.EvaluationContext{
        "user_id": "usr-42",
        "plano":   "pro",
        "pais":    "BR",
    })
    fmt.Println("valor:", result.StringValue("default"))

    // Avaliar todos os flags
    all, _ := client.EvaluateAll(ctx, nil)
    theme := all.Get("ui-theme").StringValue("light")
}
```

### Opções do cliente

| Opção | Descrição |
|-------|-----------|
| `sdk.WithTimeout(d)` | Timeout para chamadas HTTP |
| `sdk.WithHTTPClient(hc)` | Cliente HTTP customizado |

### Tipos de valor

`BoolValue(default)` · `StringValue(default)` · `Float64Value(default)` · `IntValue(default)` · `JSONValue(target)`

---

## Dashboard — Páginas

| Página | Descrição |
|--------|-----------|
| Dashboard | Estatísticas e resumo do tenant |
| Applications | Gerenciar aplicações |
| Environments | Ambientes por aplicação |
| Feature Flags | Criar, editar, ativar/desativar flags |
| Targeting Rules | Regras de segmentação por flag |
| API Keys | Gerar e revogar chaves de API |
| Users | Gerenciar usuários e papéis (owner / admin / member / viewer) |
| Audit Log | Histórico de alterações |
| Usage Analytics | Métricas de avaliação de flags |
| Settings | Configurações do tenant e conta |

---

## Papéis de Usuário

| Papel | Nível | Permissões |
|-------|-------|-----------|
| `owner` | 100 | Acesso total; imutável — ninguém pode alterar ou remover |
| `admin` | 75 | Pode gerenciar `member` e `viewer`; não pode alterar `owner` ou outros `admin` |
| `member` | 50 | Acesso operacional (ler/editar flags) |
| `viewer` | 25 | Somente leitura |

---

## WebSocket em Tempo Real

```
ws://localhost:9001/api/v1/flagflash/sdk/ws?api_key=ff_...
```

**Fluxo:**
1. SDK conecta e recebe snapshot completo dos flags (`flags` message)
2. No dashboard, qualquer alteração num flag dispara evento Redis Pub/Sub
3. O WebSocket Hub entrega `flag_updated` / `flag_deleted` a todos os clientes do ambiente
4. SDK atualiza o cache local imediatamente — sem polling

**Tipos de mensagem:** `connected` · `flags` · `flag_updated` · `flag_deleted` · `ping` / `pong`

---

## Banco de Dados

| Tabela | Descrição |
|--------|-----------|
| `tenants` | Organizações/workspaces |
| `users` | Usuários do dashboard |
| `user_tenant_memberships` | Relação usuário ↔ tenant com papel |
| `applications` | Aplicações por tenant |
| `environments` | Ambientes por aplicação |
| `feature_flags` | Flags por ambiente (boolean / json / string / number) |
| `targeting_rules` | Regras de segmentação por flag |
| `api_keys` | Chaves de autenticação do SDK |
| `audit_logs` | Histórico de mudanças |
| `evaluation_events` | Eventos brutos de avaliação do SDK |
| `evaluation_summary` | Agregação horária para métricas |

As migrations ficam em `api/migrations/` e são aplicadas automaticamente pelo Docker Compose.

---

## Desenvolvimento

```bash
# Apenas infraestrutura
docker-compose up -d postgres redis

# API com live reload (requer air)
cd api && air

# Frontend
cd app && npm run dev

# Build de produção do frontend
cd app && npm run build

# Verificar types TypeScript
cd app && npx tsc --noEmit

# Build da API
cd api && go build ./...
```

---

## Licença

MIT


## Características

- 🐳 Gerenciamento completo de containers, imagens, volumes e redes
- 🖥️ Interface web React moderna com Tailwind CSS
- 🔧 API REST em Go com documentação OpenAPI
- 💻 App desktop Electron (opcional)
- 🐋 Deploy via Docker Compose ou execução local

## Modos de Execução

### 1. Docker Compose (Recomendado para produção)

```bash
# Iniciar todos os serviços
docker compose up -d

# Frontend: http://localhost:3000
# API: http://localhost:9001
```

### 2. Desenvolvimento Local (Sem Docker)

```bash
# Usando o script de desenvolvimento
./run.sh dev

# Ou usando Make
make dev

# Ou manualmente:
# Terminal 1 - API
cd api && go run ./cmd/api

# Terminal 2 - Frontend
cd app && npm run dev
```

### 3. App Desktop Electron

```bash
# Desenvolvimento
cd app
npm install
npm run dev:electron

# Build para distribuição
make electron
# ou
cd app && npm run electron:build
```

## Build do App Electron (Passo a Passo)

O app Electron inclui a API Go embutida, permitindo execução standalone sem Docker.

### 1. Compilar o binário da API

```bash
cd api
go build -o bin/api ./cmd/api
```

### 2. Instalar dependências do frontend

```bash
cd app
npm install
```

### 3. Gerar o pacote de distribuição

```bash
# Linux (gera .deb e AppImage)
npm run electron:build

# Apenas Linux
npm run electron:build:linux

# Windows
npm run electron:build:win

# macOS
npm run electron:build:mac
```

### 4. Instalar o pacote gerado

```bash
# Debian/Ubuntu (.deb)
sudo dpkg -i app/release/docker-management-ui_1.0.0_amd64.deb

# Ou usar AppImage diretamente
chmod +x "app/release/Docker Manager-1.0.0.AppImage"
./app/release/Docker\ Manager-1.0.0.AppImage
```

### Arquivos gerados

Após o build, os pacotes ficam em `app/release/`:

| Arquivo | Plataforma | Descrição |
|---------|------------|-----------|
| `docker-management-ui_1.0.0_amd64.deb` | Linux | Pacote Debian/Ubuntu |
| `Docker Manager-1.0.0.AppImage` | Linux | Executável portátil |
| `Docker Manager Setup 1.0.0.exe` | Windows | Instalador NSIS |
| `Docker Manager-1.0.0.dmg` | macOS | Instalador DMG |

### Personalização do ícone

Coloque os ícones na pasta `app/build/`:

- `icon.png` - Linux (mínimo 256x256, recomendado 512x512)
- `icon.ico` - Windows (256x256)
- `icon.icns` - macOS (512x512 ou 1024x1024)

```bash
# Converter PNG para ICO (requer ImageMagick)
convert icon.png -resize 256x256 icon.ico
```

### Como funciona a API embutida

O app Electron verifica automaticamente:

1. Se já existe uma API rodando na porta 9001 (ex: via Docker) → usa ela
2. Se não → inicia o binário embutido em `/opt/Docker Manager/resources/api/api`

Isso permite usar o app tanto com Docker quanto standalone.

## Pré-requisitos

### Para Docker Compose
- Docker e Docker Compose

### Para Desenvolvimento Local
- Go 1.24+
- Node.js 18+
- npm
- Docker (para a API gerenciar containers)

### Para Electron
- Todos os requisitos de desenvolvimento local
- Dependências do Electron Builder para sua plataforma

## Estrutura do Projeto

```
├── api/                    # Backend Go
│   ├── cmd/api/           # Ponto de entrada
│   └── internal/          # Lógica da aplicação (Clean Architecture)
├── app/                    # Frontend React
│   ├── src/               # Código fonte React
│   └── electron/          # Código Electron
├── docker-compose.yml      # Configuração Docker
├── Makefile               # Comandos de build
└── run.sh                 # Script de desenvolvimento
```

## Comandos Disponíveis

### Makefile

```bash
make help           # Mostrar ajuda
make dev            # Rodar API + Frontend (desenvolvimento)
make api            # Rodar apenas a API
make app            # Rodar apenas o frontend
make electron-dev   # Rodar app Electron em desenvolvimento
make electron       # Build do app Electron
make docker-up      # Iniciar com Docker Compose
make docker-down    # Parar Docker Compose
make clean          # Limpar artefatos de build
```

### Script run.sh

```bash
./run.sh help       # Mostrar ajuda
./run.sh dev        # Rodar API + Frontend
./run.sh api        # Rodar apenas a API
./run.sh app        # Rodar apenas o frontend
./run.sh electron   # Rodar app Electron
./run.sh build      # Build para produção
```

## Configuração

### Variáveis de Ambiente

#### API
- `PORT`: Porta da API (padrão: 9001)
- `DOCKER_HOST`: Socket do Docker (padrão: unix:///var/run/docker.sock)

#### Frontend
- `VITE_API_URL`: URL da API (padrão: /api com proxy)

## Permissões do Docker

Para executar localmente, seu usuário precisa ter acesso ao Docker:

```bash
sudo usermod -aG docker $USER
newgrp docker  # ou faça logout/login
```

## Licença

MIT
