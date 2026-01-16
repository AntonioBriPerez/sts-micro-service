# ETAPA 1: BUILDER (Compilación)
# Usamos la imagen oficial completa para tener el compilador
FROM golang:1.23 AS builder

WORKDIR /app

# Copiamos gestión de dependencias primero (Caché de capas de Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el código fuente
COPY main.go .

# Compilamos un binario estático (CGO_ENABLED=0 es vital para Alpine/Distroless)
RUN CGO_ENABLED=0 GOOS=linux go build -o sts-server main.go

# ---------------------------------------------------------

# ETAPA 2: RUNNER (Ejecución)
# Usamos Alpine Linux (super ligero, ~5MB)
FROM alpine:latest

WORKDIR /root/

# Instalamos certificados CA por si el STS necesita llamar a fuera (buenas prácticas)
RUN apk --no-cache add ca-certificates

# Copiamos SOLO el binario compilado desde la etapa 1
COPY --from=builder /app/sts-server .

# Documentamos el puerto
EXPOSE 8080

# Comando de arranque
CMD ["./sts-server"]
