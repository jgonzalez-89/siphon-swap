# 🚀 CryptoSwap MVP

Exchange aggregator minimalista con HTMX + Go

## ✨ Features

- ✅ 3 exchanges reales (ChangeNOW, SimpleSwap, StealthEX)
- ✅ Comparación de rates en tiempo real
- ✅ Sin JavaScript (HTMX)
- ✅ Interfaz moderna y minimalista
- ✅ Cache inteligente
- ✅ Arquitectura escalable

## 📁 Estructura

\`\`\`
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
│   └── stealthex.go       
│
├── models/                # Tipos compartidos
│   └── types.go           
│
└── static/                # Frontend
    └── index.html         
\`\`\`

## 🚀 Quick Start

### 1. Clonar y configurar

\`\`\`bash
# Crear estructura
mkdir -p cryptoswap/{handlers,services,exchanges,models,static}
cd cryptoswap

# Crear archivos (copiar el código de cada archivo)
touch main.go .env go.mod
touch models/types.go
touch exchanges/{changenow,simpleswap,stealthex}.go
touch services/{aggregator,cache}.go
touch handlers/{quote,currencies,swap}.go
touch static/index.html
\`\`\`

### 2. Configurar API Keys

Editar `.env` con tus keys:

\`\`\`env
CHANGENOW_API_KEY=tu_key_aqui
SIMPLESWAP_API_KEY=tu_key_aqui
STEALTHEX_API_KEY=tu_key_aqui
\`\`\`

### 3. Instalar dependencias

\`\`\`bash
go mod init cryptoswap
go get github.com/gorilla/mux github.com/joho/godotenv
\`\`\`

### 4. Ejecutar

\`\`\`bash
go run main.go
\`\`\`

Abrir http://localhost:8080

## 📊 API Endpoints

### Get Currencies
\`\`\`bash
curl http://localhost:8080/api/currencies
\`\`\`

### Get Quote
\`\`\`bash
curl "http://localhost:8080/api/quote?from=btc&to=eth&amount=0.1"
\`\`\`

### Create Swap
\`\`\`bash
curl -X POST http://localhost:8080/api/swap \\
  -H "Content-Type: application/json" \\
  -d '{
    "from": "btc",
    "to": "eth",
    "amount": 0.1,
    "to_address": "0x...",
    "exchange": "ChangeNOW"
  }'
\`\`\`

## 🔧 Desarrollo

### Hot Reload
\`\`\`bash
go install github.com/cosmtrek/air@latest
air
\`\`\`

### Tests
\`\`\`bash
go test ./...
\`\`\`

### Build
\`\`\`bash
go build -o cryptoswap
./cryptoswap
\`\`\`

## 🐳 Docker

\`\`\`dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o cryptoswap main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/cryptoswap .
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./cryptoswap"]
\`\`\`

\`\`\`bash
docker build -t cryptoswap .
docker run -p 8080:8080 --env-file .env cryptoswap
\`\`\`

## 📈 Próximos Pasos

- [ ] WebSockets para precios real-time
- [ ] Más exchanges (1inch, Uniswap)
- [ ] Histórico de transacciones
- [ ] Redis para cache distribuido
- [ ] Tests automatizados
- [ ] CI/CD con GitHub Actions

## 📝 Licencia

MIT
\`\`\`

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