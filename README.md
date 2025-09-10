# 🚀 CryptoSwap MVP

Exchange aggregator minimalista con HTMX + Go

## ✨ Features

- ✅ 4 exchanges reales (ChangeNOW, SimpleSwap, StealthEX, LetsExchange)
- ✅ Comparación de rates en tiempo real
- ✅ Sin JavaScript (HTMX)
- ✅ Interfaz moderna y minimalista
- ✅ Cache inteligente
- ✅ Arquitectura escalable

## 📁 Estructura

\\\
cryptoswap/
├── main.go                 # Entry point
├── .env                    # API keys (no commitear!)
├── go.mod                  # Dependencias
│
├── handlers/               # HTTP handlers
│   ├── quote.go           # Cotizaciones
│   ├── swap.go            # Intercambios
│   └── currencies.go      # Monedas
│
├── services/              # Lógica de negocio
│   ├── aggregator.go      # Coordinador de exchanges
│   └── cache.go           # Cache en memoria
│
├── exchanges/             # Clientes API
│   ├── changenow.go
│   ├── simpleswap.go
│   ├── stealthex.go
│   └── letsexchange.go
│
├── models/                # Tipos compartidos
│   └── types.go
│
└── static/                # Frontend
    └── index.html
\\\

## 🚀 Quick Start

### 1. Clonar y configurar

```bash
# Crear estructura
git clone
```

### 2. Configurar API Keys

Editar `.env` con tus keys:

```env
CHANGENOW_API_KEY=tu_key_aqui
SIMPLESWAP_API_KEY=tu_key_aqui
STEALTHEX_API_KEY=tu_key_aqui
LETSEXCHANGE_API_KEY=tu_key_aqui
```

### 3. Instalar dependencias

```bash
go mod init cryptoswap
go get github.com/gorilla/mux github.com/joho/godotenv
```

### 4. Ejecutar

```bash
go run main.go
```

Abrir http://localhost:8080

## 📊 API Endpoints

### Get Currencies
```bash
curl http://localhost:8080/api/currencies
```

### Get Quote
```bash
curl "http://localhost:8080/api/quote?from=btc&to=eth&amount=0.1"
```

### Create Swap
```bash
curl -X POST http://localhost:8080/api/swap \\
  -H "Content-Type: application/json" \\
  -d '{
    "from": "btc",
    "to": "eth",
    "amount": 0.1,
    "to_address": "0x...",
    "exchange": "ChangeNOW"
  }'
```

## 🔧 Desarrollo
### Tests
```bash
go test ./...
```

### Docker Build
```bash
make docker.build
```

## 📈 Próximos Pasos

- [ ] WebSockets para precios real-time
- [ ] Más exchanges (1inch, Uniswap)
- [ ] Histórico de transacciones
- [ ] Redis para cache distribuido
- [ ] Tests automatizados
- [ ] CI/CD con GitHub Actions

## 📝 Licencia

MIT


## `Makefile` - Comandos útiles

```makefile
.PHONY: help
help: ## Mostrar ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Ejecutar servidor
	go run main.go

.PHONY: dev
dev: ## Ejecutar con hot reload
	air

.PHONY: build
build: ## Compilar binario
	go build -o bin/cryptoswap main.go

.PHONY: test
test: ## Ejecutar tests
	go test -v ./...

.PHONY: clean
clean: ## Limpiar archivos generados
	rm -rf bin/

.PHONY: deps
deps: ## Instalar dependencias
	go mod download
	go mod tidy

.PHONY: fmt
fmt: ## Formatear código
	go fmt ./...

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t cryptoswap:latest .

.PHONY: docker-run
docker-run: ## Run con Docker
	docker run -p 8080:8080 --env-file .env cryptoswap:latest
```
