# Inclui o arquivo .env se ele existir para que as variáveis fiquem disponíveis no Makefile
include .env
export

.PHONY: migrate
migrate:
	@echo "Rodando as migrations com Tern..."
	cd internal/store/pgstore/migrations && tern migrate --config ./tern.conf --migrations .

air:
	air --build.cmd "go build -o ./tmp/main ./cmd/api/main.go" --build.bin "./tmp/main"
