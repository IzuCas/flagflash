# FlagFlash — Arquitetura

## 📐 Visão Geral

```
┌──────────────────────────────────────────────────────────────────┐
│                       FlagFlash Platform                         │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────┐   ┌─────────────────┐   ┌───────────────┐  │
│  │  Web Dashboard  │   │   SDK (Go)      │   │  REST Client  │  │
│  │  React / TS     │   │  gorilla/ws     │   │  qualquer     │  │
│  └────────┬────────┘   └────────┬────────┘   └──────┬────────┘  │
│           │  JWT                │  API Key          │  API Key  │
│           └────────────────────┬┴───────────────────┘           │
│                                │                                 │
│               ┌────────────────▼───────────────┐                │
│               │       API Gateway (Go)          │                │
│               │  Chi router + Huma v2 (OpenAPI) │                │
│               │  ├─ /auth/*   JWT público       │                │
│               │  ├─ /sdk/*    API Key           │                │
│               │  ├─ /sdk/ws   WebSocket         │                │
│               │  └─ /manage/* JWT dashboard     │                │
│               └────────────────┬───────────────┘                │
│                                │                                 │
│        ┌───────────────────────┼───────────────┐                │
│        │                       │               │                │
│  ┌─────▼──────┐   ┌────────────▼────┐   ┌──────▼──────┐        │
│  │ Application│   │    Domain       │   │Infrastructure│        │
│  │  Services  │   │  Entities       │   │  PostgreSQL  │        │
│  │  (10 svcs) │   │  Repositories   │   │  Redis       │        │
│  └────────────┘   └─────────────────┘   │  WebSocket   │        │
│                                         └──────────────┘        │
└──────────────────────────────────────────────────────────────────┘
```

## 🏗️ Hierarquia de Domínio

```
Tenant
├── Users (memberships: owner / admin / member / viewer)
├── Invite Tokens (convites pendentes por email)
├── Applications
│   └── Environments (dev, staging, production)
│       ├── Feature Flags (boolean / json / string / number)
│       │   └── Targeting Rules (condições + rollout %)
│       └── API Keys (escopo por environment)
└── Audit Logs
```

### Papéis de Usuário

| Papel | Nível | Pode gerenciar |
|-------|-------|----------------|
| `owner` | 100 | Todos; imutável — ninguém pode alterar ou remover |
| `admin` | 75 | `member` e `viewer` |
| `member` | 50 | Acesso operacional (flags, environments) |
| `viewer` | 25 | Somente leitura |

## 📦 Estrutura de Pastas

```
flagflash/
├── docker-compose.yml             # PostgreSQL + Redis
├── api/                           # Backend Go (Clean Architecture)
│   ├── cmd/api/main.go            # Ponto de entrada
│   ├── internal/
│   │   ├── domain/                # Entidades + interfaces de repositório
│   │   │   ├── entity/
│   │   │   │   ├── tenant.go
│   │   │   │   ├── user.go        # UserRole + hierarquia
│   │   │   │   ├── application.go
│   │   │   │   ├── environment.go
│   │   │   │   ├── feature_flag.go
│   │   │   │   ├── targeting.go
│   │   │   │   ├── api_key.go
│   │   │   │   ├── audit_log.go
│   │   │   │   └── evaluation_event.go
│   │   │   └── repository/        # Interfaces (10 repositories)
│   │   ├── application/service/   # Serviços de negócio (10 services)
│   │   │   ├── auth_service.go
│   │   │   ├── tenant_service.go
│   │   │   ├── user_service.go
│   │   │   ├── application_service.go
│   │   │   ├── environment_service.go
│   │   │   ├── feature_flag_service.go
│   │   │   ├── api_key_service.go
│   │   │   ├── evaluation_service.go
│   │   │   ├── audit_log_service.go
│   │   │   └── usage_metrics_service.go
│   │   ├── infrastructure/
│   │   │   ├── config/config.go
│   │   │   ├── email/email.go       # Serviço SMTP para e-mails de convite
│   │   │   ├── postgres/          # Implementações dos repositórios
│   │   │   ├── redis/             # Cache (flag_cache, apikey_cache, etc.)
│   │   │   └── websocket/hub.go   # WebSocket Hub
│   │   └── interfaces/http/
│   │       ├── flagflash_router.go
│   │       ├── handler/           # 9 handlers HTTP
│   │       └── middleware/        # JWT auth, API Key auth, logging
│   ├── migrations/001_initial_schema.sql
│   └── pkg/
│       ├── auth/                  # JWT + credentials
│       ├── middleware/            # rate_limit, logging
│       └── logger/
├── app/                           # Frontend React + TS + Tailwind CSS
│   └── src/
│       ├── pages/flagflash/       # 12 páginas do dashboard
│       ├── components/            # Modal, ConfirmDeleteModal
│       ├── contexts/AuthContext.tsx
│       ├── services/flagflash-api.ts
│       ├── hooks/useWebSocket.ts
│       └── types/flagflash.ts
├── sdk/                           # Go SDK
│   ├── client.go
│   ├── evaluate.go
│   ├── websocket.go
│   └── example/basic/main.go
└── docs/
    └── ARCHITECTURE.md
```

## 🗃️ Modelo de Dados

| Tabela | Descrição | Chaves principais |
|--------|-----------|-------------------|
| `tenants` | Organizações/workspaces | `slug` UNIQUE, `plan`, `active`, `settings` JSONB |
| `users` | Usuários do dashboard | `email` UNIQUE, `password_hash`, `role`, `active` |
| `user_tenant_memberships` | N-N usuários ↔ tenants | `(user_id, tenant_id)` UNIQUE, `role`, `active` |
| `applications` | Apps por tenant | `(tenant_id, slug)` UNIQUE |
| `environments` | Ambientes por app | `(application_id, slug)` UNIQUE, `color`, `is_production` |
| `feature_flags` | Flags por environment | `(environment_id, key)` UNIQUE, `type`, `value` JSONB, `tags[]` |
| `targeting_rules` | Regras por flag | `priority`, `conditions` JSONB, `percentage` 0-100 |
| `api_keys` | Autenticação SDK | `key_hash` UNIQUE, `permissions[]`, `expires_at`, `revoked_at` |
| `audit_logs` | Histórico completo | `entity_type`, `old_value`/`new_value` JSONB |
| `evaluation_events` | Eventos brutos do SDK | `flag_key`, `value`, `context` JSONB, `evaluated_at` |
| `evaluation_summary` | Agregação horária | `(env_id, flag_id, hour_bucket)` UNIQUE |
| `invite_tokens` | Tokens de convite | `token` UNIQUE, `email`, `tenant_id`, `role`, `expires_at`, `accepted_at` |

**Seed padrão:**
- Tenant `00000000-...-0001` — "Default Organization" (slug: `default`)
- Usuário `admin@flagflash.io` / `admin123` com papel `owner`

## 🔌 Rotas da API

Prefixo base: `/api/v1/flagflash`

### Autenticação (público)

```
POST  /auth/login            # Email + senha → JWT
POST  /auth/register         # Cria tenant + usuário owner
POST  /auth/refresh          # Renova JWT
POST  /auth/switch-tenant    # Troca tenant ativo no token
POST  /auth/change-password  # Altera senha do usuário logado
GET   /auth/invite/{token}   # Valida token de convite (público)
POST  /auth/invite/accept    # Aceita convite e define senha (público)
```

### SDK — API Key obrigatória

```
GET   /sdk/flags             # Todos os flags do environment
GET   /sdk/flags/{key}       # Flag por chave
POST  /sdk/evaluate          # Avalia um flag com EvaluationContext
POST  /sdk/evaluate-all      # Avalia todos os flags com contexto
GET   /sdk/ws                # WebSocket (upgrade)
```

### Dashboard `/manage` — JWT obrigatório

```
# Tenants
GET    /manage/tenants
GET    /manage/tenants/{id}
POST   /manage/tenants
PUT    /manage/tenants/{id}
DELETE /manage/tenants/{id}

# Applications
GET    /manage/tenants/{tid}/applications
POST   /manage/tenants/{tid}/applications
GET    /manage/tenants/{tid}/applications/{id}
PUT    /manage/tenants/{tid}/applications/{id}
DELETE /manage/tenants/{tid}/applications/{id}

# Environments
GET    /manage/applications/{aid}/environments
POST   /manage/applications/{aid}/environments
PUT    /manage/applications/{aid}/environments/{id}
DELETE /manage/applications/{aid}/environments/{id}

# Feature Flags
GET    /manage/environments/{eid}/flags
POST   /manage/environments/{eid}/flags
PUT    /manage/environments/{eid}/flags/{id}
PATCH  /manage/environments/{eid}/flags/{id}/toggle
POST   /manage/environments/{eid}/flags/{id}/copy
DELETE /manage/environments/{eid}/flags/{id}

# Targeting Rules
GET    /manage/flags/{fid}/targeting-rules
POST   /manage/flags/{fid}/targeting-rules
PUT    /manage/flags/{fid}/targeting-rules/{id}
DELETE /manage/flags/{fid}/targeting-rules/{id}

# API Keys
GET    /manage/tenants/{tid}/api-keys
POST   /manage/tenants/{tid}/api-keys
DELETE /manage/tenants/{tid}/api-keys/{id}
POST   /manage/tenants/{tid}/api-keys/{id}/rotate

# Users
GET    /manage/tenants/{tid}/users
POST   /manage/tenants/{tid}/users
PUT    /manage/tenants/{tid}/users/{id}
DELETE /manage/tenants/{tid}/users/{id}
POST   /manage/tenants/{tid}/users/invite

# Audit Log
GET    /manage/tenants/{tid}/audit-logs

# Usage Metrics
GET    /manage/tenants/{tid}/usage-metrics
GET    /manage/tenants/{tid}/usage-metrics/timeline
GET    /manage/tenants/{tid}/usage-metrics/flags
GET    /manage/tenants/{tid}/usage-metrics/environments
```

## 📡 Fluxo WebSocket

```
URL: ws://server/api/v1/flagflash/sdk/ws
Auth: Header "Authorization: Bearer ff_<key>" ou query param ?api_key=ff_<key>

Fluxo:
1. Cliente conecta → servidor valida API Key → identifica environment
2. Servidor envia snapshot completo: { type: "flags", data: [...] }
3. SDK preenche cache local
4. Alteração via dashboard:
   └─ REST PUT /flags/{id}
      └─ Redis PUBLISH flag_updates:{env_id}
         └─ WebSocket Hub → todos os clientes do environment
            └─ SDK atualiza cache local (sem polling)

Tipos de mensagem:
  Server → Client:
    { "type": "connected", "environment_id": "..." }
    { "type": "flags",       "data": [...] }        # snapshot inicial
    { "type": "flag_updated","data": {...} }        # flag alterada
    { "type": "flag_deleted","key":  "..." }        # flag removida
    { "type": "pong" }
  Client → Server:
    { "type": "ping" }                              # keep-alive
```

## ✉️ Fluxo de Convites (Invite)

```
1. Admin/Owner convida email via POST /manage/tenants/{tid}/users/invite
2. Backend gera token seguro (32 bytes hex, crypto/rand), salva em invite_tokens (TTL 7 dias)
3. Se SMTP configurado → envia email com link: {APP_URL}/accept-invite?token={token}
   Se SMTP não configurado → retorna link no response para copiar manualmente
4. Usuário acessa link → frontend chama GET /auth/invite/{token} para validar
5. Usuário preenche nome + senha → frontend chama POST /auth/invite/accept
6. Backend cria usuário (se novo) + cria membership no tenant + marca token como aceito
7. Usuário é redirecionado para login
```

**Variáveis de configuração SMTP:**

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `SMTP_HOST` | `""` | Host SMTP (vazio = envio desabilitado) |
| `SMTP_PORT` | `587` | Porta SMTP |
| `SMTP_USERNAME` | `""` | Usuário para autenticação |
| `SMTP_PASSWORD` | `""` | Senha para autenticação |
| `SMTP_FROM` | `""` | Email remetente |
| `APP_URL` | `http://localhost:5173` | URL base do frontend (usada nos links) |

## 🔐 Estratégia de Cache (Redis)

```
Chaves de cache:
  flag:{env_id}:{flag_key}   → Flag individual (JSON)     TTL: 5 min
  flag:list:{env_id}         → Lista de flags do ambiente  TTL: 5 min
  apikey:{key_hash}          → Dados da API Key            TTL: 1 hora
  tenant:{id}                → Dados do tenant             TTL: 30 min
  tenant:slug:{slug}         → Lookup por slug             TTL: 30 min
  app:{id}                   → Dados da application        TTL: 5 min
  app:list:{tenant_id}       → Lista de apps               TTL: 5 min
  env:{id}                   → Dados do environment        TTL: 5 min
  env:list:{app_id}          → Lista de environments       TTL: 5 min

Invalidação:
  Flag alterada
    → redis PUBLISH flag_updates:{env_id}
    → WebSocket Hub recebe
    → Entrega flag_updated a todos SDKs do environment
    → Invalida cache da flag

Caches implementados (arquivos separados):
  flag_cache.go / apikey_cache.go / tenant_cache.go
  application_cache.go / environment_cache.go
```

## 🚀 Escalabilidade

- **API stateless** — escala horizontalmente; múltiplas instâncias compartilham o Redis
- **WebSocket** — Redis Pub/Sub sincroniza eventos entre instâncias
- **Banco** — connection pool configurável (`DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`)
- **Cache em camadas** — Redis (L2) reduz carga no Postgres; SDK mantém L1 em memória
- **Rate limiting** — 300 req/min por IP (middleware global)
- **Avaliação de flags** — toda a lógica de targeting roda no servidor; SDK só decide pelo cache local

## 🎯 Avaliação de Flags com Targeting

```
Ordem de avaliação (POST /sdk/evaluate):
  1. Busca flag do cache Redis
  2. Flag desabilitada → retorna default_value
  3. Para cada targeting rule (ordenado por priority ASC):
     a. Avalia todas as conditions (operadores: eq, neq, in, not_in,
        contains, starts_with, ends_with, gt, lt, gte, lte, regex)
     b. Conditions satisfeitas → verifica rollout percentage (hash do user_id)
     c. Aprovado → retorna value da rule
  4. Nenhuma rule faz match → retorna default_value

EvaluationContext (mapa livre de chaves/valores):
  {
    "user_id":  "usr-42",
    "email":    "user@company.com",
    "plan":     "pro",
    "country":  "BR",
    "version":  "2.1.0"
  }
```

## 📊 Exemplo de Targeting Rule

```json
[
  {
    "name": "Beta Users",
    "priority": 1,
    "enabled": true,
    "conditions": [
      { "attribute": "email", "operator": "ends_with", "value": "@company.com" }
    ],
    "value": true,
    "percentage": 100
  },
  {
    "name": "Gradual Rollout BR/US",
    "priority": 2,
    "enabled": true,
    "conditions": [
      { "attribute": "country", "operator": "in", "value": ["BR", "US"] }
    ],
    "value": true,
    "percentage": 25
  }
]
```

## 🏃 Como Rodar

```bash
# 1. Infraestrutura
docker-compose up -d

# 2. API
cd api
export JWT_SECRET="min-32-chars-secret"
go run ./cmd/api

# 3. Frontend
cd app && npm install && npm run dev

# Endereços:
#   Dashboard  → http://localhost:5173
#   API        → http://localhost:9001/api/v1/flagflash
#   OpenAPI    → http://localhost:9001/api/v1/flagflash/docs
#   Health     → http://localhost:9001/health
#
# Login padrão: admin@flagflash.io / admin123
```

## 🔒 Segurança

- **JWT** algoritmo fixado em HS256; app recusa token com alg diferente
- **JWT_SECRET** obrigatório no startup (panic se ausente ou < 32 chars)
- **Owners** são imutáveis — nenhum papel pode editar ou remover um owner
- **Hierarquia de papéis** validada na API antes de qualquer mutação de usuário
- **Rate limiting** 300 req/min por IP via `x/time/rate`
- **CORS** configurável via `CORS_ALLOWED_ORIGINS`; credenciais não expostas
- **SSL Postgres** default `prefer`
- **API Keys** armazenadas como hash SHA-256; nunca o valor original
- **Invite Tokens** expiram em 7 dias; token de 32 bytes hex gerado com `crypto/rand`
- **Paginação** com LIMIT/OFFSET sempre parametrizado (sem SQL injection)
- **SDK** valida formato da flag key com regex antes de qualquer chamada
