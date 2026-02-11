# API Orders - Product Management

API backend escalável construída com Go, Gin Framework, MongoDB e RabbitMQ seguindo Arquitetura Hexagonal (Clean Architecture).

## Tecnologias

- **Go 1.25.3**
- **Gin** - Framework HTTP
- **MongoDB** - Banco de dados NoSQL
- **RabbitMQ** - Message broker
- **Wire** - Dependency Injection
- **Zap** - Structured logging
- **Validator v10** - Validação de dados
- **Swagger/OpenAPI** - Documentação interativa da API

## Arquitetura

O projeto segue **Arquitetura Hexagonal** (Ports & Adapters) com separação clara de responsabilidades:

```
internal/
├── core/               # Camada de Domínio (Lógica de negócio)
│   ├── domain/         # Entidades do domínio
│   ├── dto/            # Data Transfer Objects
│   ├── ports/          # Interfaces (Repository, UseCase)
│   └── usecase/        # Casos de uso
│
├── adapter/            # Camada de Adaptadores
│   ├── http/           # Handlers, Routes, Middlewares
│   ├── repository/     # Implementações de repositórios
│   └── messages/       # Message producers (RabbitMQ)
│
└── infra/              # Infraestrutura
    └── database/       # Conexões de banco de dados
```

## Serviços

### 1. API (api-orders)
Serviço responsavel pelo cadastro de produtos, geração da ordem, visualização de uma ordem criada, atualização dos status de uma ordem.


### 2. Manager Status

Serviço consumidor. Quando uma nova ordem é criada, uma mensagem é publicada na fila do rabbitMQ, esse serviço consome essa mensagem e atualiza o status da ordem de criada para em_processamento. Quando o status de uma ordem é atualizado, uma mensagem é gerada na fila e esse serviço atualiza o status da ordem.

### 3. Instruçoess de uso
Nos 2 diretórios (api-orders, manager-status), incluir sua senha do mongodb atlas no arquivo de configuração config.toml. Depois basta executar o docker compose

```bash
docker compose up -d
```


## Documentação Interativa (Swagger)

A API possui documentação interativa completa usando Swagger/OpenAPI.

**Acessar Swagger UI:**

```
http://localhost:8000/swagger/index.html
```


## Endpoints

### Health Check

```bash
GET /health
```

**Description:** Verifica o status da API e da conexão com RabbitMQ.

**Success Response (200 OK):**
```json
{
  "api_status": "ok",
  "rabbitmq_status": "sucesso"
}
```

**Quando RabbitMQ está indisponível:**
```json
{
  "api_status": "ok",
  "rabbitmq_status": "falha"
}
```

**Status possíveis:**
- `api_status`: sempre `"ok"` (se a API está respondendo)
- `rabbitmq_status`: `"sucesso"` ou `"falha"`

### Criar Produto

```bash
POST /api/v1/products
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Notebook Dell",
  "description": "Notebook Dell Inspiron 15 com 16GB RAM",
  "quantity": 10,
  "price": 3999.99
}
```

**Validações:**
- `name`: obrigatório, mínimo 3 caracteres
- `description`: obrigatório
- `quantity`: obrigatório, >= 0
- `price`: obrigatório, > 0

**Success Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "_id": "507f1f77bcf86cd799439011",
    "name": "Notebook Dell",
    "description": "Notebook Dell Inspiron 15 com 16GB RAM",
    "quantity": 10,
    "price": 3999.99,
    "created_at": "2024-02-10T12:00:00Z",
    "updated_at": "2024-02-10T12:00:00Z"
  },
  "message": "Product created successfully"
}
```

## Orders (Pedidos)

### Criar Pedido

```bash
POST /api/v1/orders
Content-Type: application/json
```

**Request Body:**
```json
{
  "items": [
    {
      "product_id": "698c0a0893c94ce530171bbb",
      "quantity": 6
    }
  ]
}
```

**Validações:**
- `items`: obrigatório, mínimo 1 item
- `items[].product_id`: obrigatório, deve ser um ObjectID válido
- `items[].quantity`: obrigatório, >= 1
- O produto deve existir no banco de dados
- Deve haver estoque suficiente para a quantidade solicitada

**Comportamento:**
- Busca automaticamente os detalhes do produto (nome e preço) pelo `product_id`
- Valida disponibilidade de estoque
- Calcula o total do pedido
- Publica mensagem no RabbitMQ (fila `order-status`)
- Registra a publicação no MongoDB (`published_orders`)

**Success Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "_id": "67ab3f2d8c9e1a2b3c4d5e6f",
    "order_number": "ORD-A1B2C3D4",
    "items": [
      {
        "product_id": "698c0a0893c94ce530171bbb",
        "product_name": "Mouse Gamer",
        "price": 199.90,
        "quantity": 6
      }
    ],
    "total": 1199.40,
    "status": "criado",
    "created_at": "2025-02-11T15:45:00Z",
    "updated_at": "2025-02-11T15:45:00Z"
  },
  "message": "Order created successfully"
}
```

**Error Responses:**
- `404 Not Found`: Produto não encontrado
- `400 Bad Request`: Estoque insuficiente
- `400 Bad Request`: Validação falhou (campos obrigatórios)

### Buscar Pedido por ID

```bash
GET /api/v1/orders/:id
```

**Path Parameters:**
- `id`: ObjectID do pedido (exemplo: `67ab3f2d8c9e1a2b3c4d5e6f`)

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "_id": "67ab3f2d8c9e1a2b3c4d5e6f",
    "order_number": "ORD-A1B2C3D4",
    "items": [
      {
        "product_id": "698c0a0893c94ce530171bbb",
        "product_name": "Mouse Gamer",
        "price": 199.90,
        "quantity": 6
      }
    ],
    "total": 1199.40,
    "status": "criado",
    "created_at": "2025-02-11T15:45:00Z",
    "updated_at": "2025-02-11T15:45:00Z"
  },
  "message": "Order retrieved successfully"
}
```

**Error Responses:**
- `404 Not Found`: Pedido não encontrado
- `400 Bad Request`: ID inválido

### Atualizar Status do Pedido

```bash
PATCH /api/v1/orders/:id/status
Content-Type: application/json
```

**Path Parameters:**
- `id`: ObjectID do pedido (exemplo: `67ab3f2d8c9e1a2b3c4d5e6f`)

**Request Body:**
```json
{
  "status": "enviado"
}
```

**Validações:**
- `status`: obrigatório, deve ser um dos valores: `criada`, `em_processamento`, `enviado`, `entregue`

**Comportamento:**
- Atualiza o status do pedido no MongoDB
- Publica mensagem no RabbitMQ (fila `order-status`) com o novo status
- Registra a publicação no MongoDB (`published_orders`)

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "_id": "67ab3f2d8c9e1a2b3c4d5e6f",
    "order_number": "ORD-A1B2C3D4",
    "items": [
      {
        "product_id": "698c0a0893c94ce530171bbb",
        "product_name": "Mouse Gamer",
        "price": 199.90,
        "quantity": 6
      }
    ],
    "total": 1199.40,
    "status": "enviado",
    "created_at": "2025-02-11T15:45:00Z",
    "updated_at": "2025-02-11T16:30:00Z"
  },
  "message": "Order status updated successfully"
}
```

**Error Responses:**
- `404 Not Found`: Pedido não encontrado
- `400 Bad Request`: Status inválido
- `400 Bad Request`: Campo obrigatório ausente

**Mensagem RabbitMQ Publicada:**
```json
{
  "order_id": "67ab3f2d8c9e1a2b3c4d5e6f",
  "ts": 1739287927.98399,
  "status": "enviado"
}
```




