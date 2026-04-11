# FlagFlash - Plataforma de Feature Flags

## 📐 Arquitetura Geral

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              FlagFlash Platform                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐          │
│  │   Web Dashboard  │    │    SDK Client    │    │   REST API       │          │
│  │   (React/TS)     │    │   (Go/JS/etc)    │    │   Consumer       │          │
│  └────────┬─────────┘    └────────┬─────────┘    └────────┬─────────┘          │
│           │                       │                       │                     │
│           └───────────────────────┼───────────────────────┘                     │
│                                   │                                              │
│                    ┌──────────────▼──────────────┐                              │
│                    │      API Gateway (Go)       │                              │
│                    │   - REST Endpoints          │                              │
│                    │   - WebSocket Server        │                              │
│                    │   - API Key Auth            │                              │
│                    │   - JWT Auth (Dashboard)    │                              │
│                    └──────────────┬──────────────┘                              │
│                                   │                                              │
│           ┌───────────────────────┼───────────────────────┐                     │
│           │                       │                       │                     │
│  ┌────────▼─────────┐   ┌────────▼─────────┐   ┌────────▼─────────┐           │
│  │   Application    │   │     Domain       │   │  Infrastructure  │           │
│  │     Layer        │   │     Layer        │   │     Layer        │           │
│  │   - Services     │   │   - Entities     │   │   - PostgreSQL   │           │
│  │   - Use Cases    │   │   - Repos        │   │   - Redis        │           │
│  │   - DTOs         │   │   - Value Obj    │   │   - WebSocket    │           │
│  └──────────────────┘   └──────────────────┘   └──────────────────┘           │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## 🏗️ Hierarquia de Domínio

```
Tenant (Organização)
├── Applications (Múltiplas aplicações)
│   ├── Environments (dev, staging, prod)
│   │   ├── Feature Flags
│   │   │   ├── Boolean Flags
│   │   │   ├── JSON Config Flags
│   │   │   └── Rollout Rules
│   │   └── API Keys (vinculadas ao environment)
│   └── Environments...
└── Applications...
```

## 📦 Estrutura de Pastas

```
flagflash/
├── api/                           # Backend Go
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── domain/                # Camada de Domínio
│   │   │   ├── entity/
│   │   │   │   ├── tenant.go
│   │   │   │   ├── application.go
│   │   │   │   ├── environment.go
│   │   │   │   ├── feature_flag.go
│   │   │   │   ├── api_key.go
│   │   │   │   ├── audit_log.go
│   │   │   │   └── targeting.go
│   │   │   └── repository/
│   │   │       ├── tenant_repository.go
│   │   │       ├── application_repository.go
│   │   │       ├── environment_repository.go
│   │   │       ├── feature_flag_repository.go
│   │   │       └── api_key_repository.go
│   │   ├── application/           # Camada de Aplicação
│   │   │   └── service/
│   │   │       ├── tenant_service.go
│   │   │       ├── application_service.go
│   │   │       ├── environment_service.go
│   │   │       ├── feature_flag_service.go
│   │   │       ├── api_key_service.go
│   │   │       └── evaluation_service.go
│   │   ├── infrastructure/        # Camada de Infraestrutura
│   │   │   ├── postgres/
│   │   │   │   ├── connection.go
│   │   │   │   ├── tenant_repo.go
│   │   │   │   ├── application_repo.go
│   │   │   │   ├── environment_repo.go
│   │   │   │   ├── feature_flag_repo.go
│   │   │   │   └── api_key_repo.go
│   │   │   ├── redis/
│   │   │   │   ├── connection.go
│   │   │   │   ├── cache.go
│   │   │   │   └── pubsub.go
│   │   │   └── websocket/
│   │   │       └── hub.go
│   │   └── interfaces/            # Camada de Interface
│   │       └── http/
│   │           ├── dto/
│   │           │   ├── tenant_dto.go
│   │           │   ├── application_dto.go
│   │           │   ├── environment_dto.go
│   │           │   ├── feature_flag_dto.go
│   │           │   └── api_key_dto.go
│   │           ├── handler/
│   │           │   ├── tenant_handler.go
│   │           │   ├── application_handler.go
│   │           │   ├── environment_handler.go
│   │           │   ├── feature_flag_handler.go
│   │           │   ├── api_key_handler.go
│   │           │   ├── evaluation_handler.go
│   │           │   └── ws_handler.go
│   │           └── router.go
│   ├── pkg/
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   └── api_key.go
│   │   ├── middleware/
│   │   │   ├── jwt_auth.go
│   │   │   ├── api_key_auth.go
│   │   │   └── tenant_context.go
│   │   └── logger/
│   │       └── logger.go
│   └── migrations/
│       └── *.sql
├── app/                           # Frontend React
│   └── src/
│       ├── pages/
│       │   ├── Tenants.tsx
│       │   ├── Applications.tsx
│       │   ├── Environments.tsx
│       │   ├── FeatureFlags.tsx
│       │   ├── ApiKeys.tsx
│       │   └── Dashboard.tsx
│       ├── components/
│       │   ├── FlagEditor.tsx
│       │   ├── JsonEditor.tsx
│       │   ├── RolloutConfig.tsx
│       │   └── TargetingRules.tsx
│       └── services/
│           └── flagflash-api.ts
└── docker-compose.yml
```

## 🗃️ Modelagem de Banco de Dados

```sql
-- Tenants (Organizações)
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

-- Applications (Aplicações por Tenant)
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    UNIQUE(tenant_id, slug)
);

-- Environments (Ambientes por Application)
CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366f1',
    is_production BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(application_id, slug)
);

-- Feature Flags
CREATE TABLE feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    flag_type VARCHAR(20) NOT NULL CHECK (flag_type IN ('boolean', 'json', 'string', 'number')),
    enabled BOOLEAN DEFAULT FALSE,
    default_value JSONB NOT NULL,
    version INTEGER DEFAULT 1,
    tags TEXT[],
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    UNIQUE(environment_id, key)
);

-- Targeting Rules (Regras de Targeting)
CREATE TABLE targeting_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    priority INTEGER DEFAULT 0,
    conditions JSONB NOT NULL,
    value JSONB NOT NULL,
    percentage INTEGER DEFAULT 100 CHECK (percentage >= 0 AND percentage <= 100),
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- API Keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(64) NOT NULL UNIQUE,
    key_prefix VARCHAR(12) NOT NULL,
    permissions TEXT[] DEFAULT ARRAY['read'],
    expires_at TIMESTAMP NULL,
    last_used_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    revoked_at TIMESTAMP NULL
);

-- Audit Logs (Histórico de Alterações)
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255),
    actor_type VARCHAR(50),
    old_value JSONB,
    new_value JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Users (Usuários do Dashboard)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'member',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

-- Indexes
CREATE INDEX idx_applications_tenant_id ON applications(tenant_id);
CREATE INDEX idx_environments_application_id ON environments(application_id);
CREATE INDEX idx_feature_flags_environment_id ON feature_flags(environment_id);
CREATE INDEX idx_feature_flags_key ON feature_flags(key);
CREATE INDEX idx_targeting_rules_flag_id ON targeting_rules(feature_flag_id);
CREATE INDEX idx_api_keys_environment_id ON api_keys(environment_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_audit_logs_tenant_entity ON audit_logs(tenant_id, entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
```

## 🔌 API Endpoints

### Autenticação
```
POST   /api/v1/auth/login           # Login (retorna JWT)
POST   /api/v1/auth/register        # Registro de novo tenant/user
POST   /api/v1/auth/refresh         # Refresh JWT token
POST   /api/v1/auth/logout          # Logout
```

### Tenants
```
GET    /api/v1/tenants              # Lista tenants (admin)
GET    /api/v1/tenants/:id          # Detalhes do tenant
PUT    /api/v1/tenants/:id          # Atualiza tenant
DELETE /api/v1/tenants/:id          # Remove tenant
```

### Applications
```
GET    /api/v1/applications                    # Lista applications do tenant
POST   /api/v1/applications                    # Cria application
GET    /api/v1/applications/:id                # Detalhes do application
PUT    /api/v1/applications/:id                # Atualiza application
DELETE /api/v1/applications/:id                # Remove application
```

### Environments
```
GET    /api/v1/applications/:appId/environments           # Lista environments
POST   /api/v1/applications/:appId/environments           # Cria environment
GET    /api/v1/applications/:appId/environments/:id       # Detalhes
PUT    /api/v1/applications/:appId/environments/:id       # Atualiza
DELETE /api/v1/applications/:appId/environments/:id       # Remove
```

### Feature Flags
```
GET    /api/v1/environments/:envId/flags                  # Lista flags
POST   /api/v1/environments/:envId/flags                  # Cria flag
GET    /api/v1/environments/:envId/flags/:id              # Detalhes
PUT    /api/v1/environments/:envId/flags/:id              # Atualiza flag
PATCH  /api/v1/environments/:envId/flags/:id/toggle       # Toggle on/off
DELETE /api/v1/environments/:envId/flags/:id              # Remove flag
GET    /api/v1/environments/:envId/flags/:id/history      # Histórico
POST   /api/v1/environments/:envId/flags/copy             # Copia flags entre envs
```

### Targeting Rules
```
GET    /api/v1/flags/:flagId/targeting                    # Lista rules
POST   /api/v1/flags/:flagId/targeting                    # Cria rule
PUT    /api/v1/flags/:flagId/targeting/:id                # Atualiza rule
DELETE /api/v1/flags/:flagId/targeting/:id                # Remove rule
```

### API Keys
```
GET    /api/v1/environments/:envId/api-keys               # Lista API keys
POST   /api/v1/environments/:envId/api-keys               # Gera nova key
DELETE /api/v1/environments/:envId/api-keys/:id           # Revoga key
POST   /api/v1/environments/:envId/api-keys/:id/rotate    # Rotaciona key
```

### SDK/Client Endpoints (autenticado via API Key)
```
GET    /api/v1/sdk/flags                                  # Retorna todas flags
GET    /api/v1/sdk/flags/:key                             # Retorna flag específica
POST   /api/v1/sdk/evaluate                               # Avalia flag com contexto
```

### WebSocket
```
ws://  /ws/flags?api_key=xxx                              # Subscribe em flags
```

## 📡 Fluxo WebSocket

```
1. Cliente conecta: ws://server/ws/flags?api_key=ff_123...
2. Servidor valida API Key e identifica environment
3. Cliente se inscreve automaticamente nas flags do environment
4. Quando uma flag muda:
   - Dashboard atualiza via REST
   - Servidor publica no Redis Pub/Sub
   - Hub WebSocket recebe e distribui para clientes conectados
   
Mensagens:
→ Server: { "type": "connected", "environment": "production" }
→ Server: { "type": "flags", "data": [...] }            # Todas flags
→ Server: { "type": "flag_updated", "data": {...} }     # Flag alterada
→ Server: { "type": "flag_deleted", "key": "..." }      # Flag removida
← Client: { "type": "ping" }                            # Keep-alive
→ Server: { "type": "pong" }
```

## 🔐 Estratégia de Cache (Redis)

```
Chaves:
- flags:{env_id}             → Hash com todas flags do environment
- flags:{env_id}:{flag_key}  → Flag individual (JSONB)
- apikey:{key_hash}          → Dados da API Key
- ws:subscribers:{env_id}    → Set de conexões WebSocket

TTL:
- Flags: 5 minutos (invalidado em mudanças)
- API Keys: 1 hora
- Conexões WS: Gerenciado pelo Hub

Invalidation:
- Flag atualizada → PUBLISH flag_updates:{env_id} → Redis Pub/Sub
- WebSocket Hub recebe → Notifica clientes → Invalida cache
```

## 🚀 Estratégias de Escalabilidade

### Horizontal
- **API Server**: Stateless, escala horizontalmente
- **WebSocket**: Redis Pub/Sub para sincronizar entre instâncias
- **Database**: Read replicas para consultas

### Cache
- L1: In-memory (por instância)
- L2: Redis (compartilhado)
- Invalidação: Event-driven via Pub/Sub

### Performance
- Flags cached por environment
- Batch evaluation para múltiplas flags
- Connection pooling para PostgreSQL
- Compressão de WebSocket messages

## 🎯 Avaliação de Flags com Targeting

```go
type EvaluationContext struct {
    UserID     string         `json:"user_id"`
    Email      string         `json:"email"`
    Country    string         `json:"country"`
    Version    string         `json:"version"`
    Custom     map[string]any `json:"custom"`
}

// Processo de avaliação:
// 1. Busca flag do cache
// 2. Se flag desabilitada → retorna default
// 3. Para cada targeting rule (ordenado por prioridade):
//    a. Avalia condições
//    b. Se match, verifica percentage (rollout)
//    c. Se passa, retorna valor da rule
// 4. Se nenhuma rule match → retorna default_value
```

## 📊 Exemplo de Targeting Rules

```json
{
  "rules": [
    {
      "name": "Beta Users",
      "conditions": [
        { "attribute": "email", "operator": "ends_with", "value": "@company.com" }
      ],
      "value": true,
      "percentage": 100
    },
    {
      "name": "Gradual Rollout",
      "conditions": [
        { "attribute": "country", "operator": "in", "value": ["BR", "US"] }
      ],
      "value": true,
      "percentage": 25
    }
  ]
}
```

## 🏃 Como Rodar

```bash
# Com Docker Compose
docker-compose up -d

# Acessar:
# - Dashboard: http://localhost:3000
# - API: http://localhost:9001
# - API Docs: http://localhost:9001/docs
```
