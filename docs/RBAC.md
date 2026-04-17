# FlagFlash — Controle de Acesso (RBAC)

Este documento descreve o sistema de controle de acesso baseado em papéis (Role-Based Access Control) do FlagFlash.

## Hierarquia de Papéis

```
Owner (100)  →  Controle total do tenant
    │
Admin (75)   →  Administração e gerenciamento de usuários
    │
Member (50)  →  Acesso operacional (criar/editar recursos)
    │
Viewer (25)  →  Somente leitura
```

## Matriz de Permissões

### Gerenciamento de Usuários

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Listar usuários | ✅ | ✅ | ✅ | ✅ |
| Ver detalhes do usuário | ✅ | ✅ | ✅ | ✅ |
| Criar usuário | ❌ | ❌ | ✅ | ✅ |
| Editar usuário | ❌ | ❌ | ✅ | ✅ |
| Excluir usuário | ❌ | ❌ | ✅ | ✅ |
| Convidar usuário | ❌ | ❌ | ✅ | ✅ |
| Alterar role de usuário | ❌ | ❌ | ✅* | ✅* |

> *Admin pode gerenciar apenas `member` e `viewer`. Owner pode gerenciar `admin`, `member` e `viewer`. Ninguém pode alterar outro `owner`.

### Applications

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Listar applications | ✅ | ✅ | ✅ | ✅ |
| Ver application | ✅ | ✅ | ✅ | ✅ |
| Criar application | ❌ | ✅ | ✅ | ✅ |
| Editar application | ❌ | ✅ | ✅ | ✅ |
| Excluir application | ❌ | ❌ | ✅ | ✅ |

### Environments

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Listar environments | ✅ | ✅ | ✅ | ✅ |
| Ver environment | ✅ | ✅ | ✅ | ✅ |
| Criar environment | ❌ | ✅ | ✅ | ✅ |
| Editar environment | ❌ | ✅ | ✅ | ✅ |
| Excluir environment | ❌ | ❌ | ✅ | ✅ |
| Copiar environment | ❌ | ✅ | ✅ | ✅ |

### Feature Flags

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Listar flags | ✅ | ✅ | ✅ | ✅ |
| Ver flag | ✅ | ✅ | ✅ | ✅ |
| Criar flag | ❌ | ✅ | ✅ | ✅ |
| Editar flag | ❌ | ✅ | ✅ | ✅ |
| Toggle (ativar/desativar) | ❌ | ✅ | ✅ | ✅ |
| Excluir flag | ❌ | ❌ | ✅ | ✅ |
| Copiar flags entre envs | ❌ | ✅ | ✅ | ✅ |

### Targeting Rules

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Listar rules | ✅ | ✅ | ✅ | ✅ |
| Ver rule | ✅ | ✅ | ✅ | ✅ |
| Criar rule | ❌ | ✅ | ✅ | ✅ |
| Editar rule | ❌ | ✅ | ✅ | ✅ |
| Reordenar rules | ❌ | ✅ | ✅ | ✅ |
| Excluir rule | ❌ | ❌ | ✅ | ✅ |

### API Keys

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Listar API keys | ❌ | ✅ | ✅ | ✅ |
| Ver API key | ❌ | ✅ | ✅ | ✅ |
| Criar API key | ❌ | ❌ | ✅ | ✅ |
| Revogar API key | ❌ | ❌ | ✅ | ✅ |
| Excluir API key | ❌ | ❌ | ✅ | ✅ |

### Audit Logs

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Ver audit logs | ❌ | ❌ | ✅ | ✅ |

### Tenant Settings

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Ver configurações | ✅ | ✅ | ✅ | ✅ |
| Editar tenant | ❌ | ❌ | ❌ | ✅ |
| Excluir tenant | ❌ | ❌ | ❌ | ✅ |

### Métricas de Uso

| Ação | Viewer | Member | Admin | Owner |
|------|:------:|:------:|:-----:|:-----:|
| Ver métricas | ✅ | ✅ | ✅ | ✅ |

## Implementação

### Backend (Go)

As verificações de permissão são feitas no nível do handler usando o middleware de autorização:

```go
// Verificar se usuário tem acesso ao tenant
if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
    return nil, err
}

// Verificar se usuário é admin ou owner
if err := middleware.RequireAdminOrOwner(ctx); err != nil {
    return nil, err
}

// Verificar se usuário tem pelo menos role "member"
if err := middleware.RequireRole(ctx, "member"); err != nil {
    return nil, err
}
```

**Arquivo:** `api/internal/interfaces/http/middleware/authorization.go`

### Frontend (React/TypeScript)

O hook `usePermissions` centraliza todas as verificações de permissão:

```typescript
import { usePermissions } from '../hooks/usePermissions';

function MyComponent() {
  const { canCreateFeatureFlag, canDeleteFeatureFlag, role } = usePermissions();
  
  return (
    <div>
      {canCreateFeatureFlag && (
        <button>Create Flag</button>
      )}
      
      {canDeleteFeatureFlag && (
        <button>Delete</button>
      )}
    </div>
  );
}
```

**Arquivo:** `app/src/hooks/usePermissions.ts`

## Regras de Negócio

### Proteção de Owners

- Nenhum usuário pode alterar ou remover um `owner`
- Owners são imutáveis na hierarquia
- Apenas o sistema pode remover ou rebaixar um owner (via banco de dados direto)

### Hierarquia de Gerenciamento

Um usuário só pode gerenciar outros usuários com role **inferior** ao seu:

- **Owner** pode gerenciar: admin, member, viewer
- **Admin** pode gerenciar: member, viewer
- **Member** não pode gerenciar ninguém
- **Viewer** não pode gerenciar ninguém

### Auto-gerenciamento

- Usuários não podem alterar seu próprio role
- Usuários não podem excluir a si mesmos
- Alterações pessoais (nome, senha) são feitas via Settings

## Endpoints Protegidos

### Requer Admin ou Owner
- `POST /tenants/{id}/users` - Criar usuário
- `PUT /tenants/{id}/users/{id}` - Editar usuário
- `DELETE /tenants/{id}/users/{id}` - Excluir usuário
- `POST /tenants/{id}/users/invite` - Convidar usuário
- `PATCH /tenants/{id}/users/{id}/role` - Alterar role
- `POST /tenants/{id}/api-keys` - Criar API key
- `POST /tenants/{id}/api-keys/{id}/revoke` - Revogar API key
- `DELETE /tenants/{id}/api-keys/{id}` - Excluir API key
- `DELETE /tenants/{id}/applications/{id}` - Excluir application
- `DELETE /tenants/{id}/.../environments/{id}` - Excluir environment
- `DELETE /tenants/{id}/.../flags/{id}` - Excluir flag
- `DELETE /tenants/{id}/.../rules/{id}` - Excluir targeting rule
- `GET /tenants/{id}/audit-logs` - Ver audit logs

### Requer Member ou Superior
- `POST /tenants/{id}/applications` - Criar application
- `PUT /tenants/{id}/applications/{id}` - Editar application
- `POST /tenants/{id}/.../environments` - Criar environment
- `PUT /tenants/{id}/.../environments/{id}` - Editar environment
- `POST /tenants/{id}/.../environments/{id}/copy` - Copiar environment
- `POST /tenants/{id}/.../flags` - Criar flag
- `PUT /tenants/{id}/.../flags/{id}` - Editar flag
- `PATCH /tenants/{id}/.../flags/{id}/toggle` - Toggle flag
- `POST /tenants/{id}/.../flags/copy` - Copiar flags
- `POST /tenants/{id}/.../rules` - Criar targeting rule
- `PUT /tenants/{id}/.../rules/{id}` - Editar targeting rule
- `POST /tenants/{id}/.../rules/reorder` - Reordenar rules

### Requer Owner
- `PUT /tenants/{id}` - Editar tenant
- `DELETE /tenants/{id}` - Excluir tenant

### Acesso de Leitura (Qualquer Role)
- Todos os endpoints `GET` de listagem e detalhes
- Métricas de uso

## Segurança

### Defesa em Profundidade

1. **Frontend**: O hook `usePermissions` esconde botões/ações não autorizadas
2. **Backend Handler**: Middleware verifica role antes de processar
3. **Backend Service**: Verificações de hierarquia para gerenciamento de usuários

### Notas Importantes

- As verificações do frontend são apenas UX — a segurança real está no backend
- Nunca confie apenas no frontend para controle de acesso
- Todas as requisições passam por verificação JWT + tenant + role no backend
