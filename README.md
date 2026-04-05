# go-labs-auction
Desafio 3 do Labs Go do curso Pós Go Expert - Full Cycle

----

## fechamento automático de leilões

### descrição

Sistema de leilões em Go com fechamento automático utilizando Goroutines. Quando um leilão é criado, uma goroutine é disparada em background para monitorar o tempo e fechar o leilão automaticamente após a duração configurada. Base repo: https://github.com/devfullcycle/labs-auction-goexpert

### implementação

A funcionalidade de fechamento automático foi implementada no arquivo `internal/infra/database/auction/create_auction.go`:

- **`scheduleAuctionClose(auctionId)`**: Goroutine que aguarda a duração configurada e atualiza o status do leilão para `Completed` no MongoDB.
- **`getAuctionDuration()`**: Lê a variável de ambiente `AUCTION_DURATION` para determinar a duração do leilão.
- A goroutine é iniciada imediatamente após a inserção do leilão no banco, sem bloquear a thread principal.

### variáveis de ambiente

Configuráveis no arquivo `cmd/auction/.env`:

| Variável | Descrição | Valor Padrão |
|---|---|---|
| `AUCTION_DURATION` | Duração do leilão antes do fechamento automático | `30s` |
| `AUCTION_INTERVAL` | Intervalo usado na validação de bids | `20s` |
| `BATCH_INSERT_INTERVAL` | Intervalo de inserção em lote de bids | `20s` |
| `MAX_BATCH_SIZE` | Tamanho máximo do lote de bids | `4` |
| `MONGODB_URL` | URL de conexão com MongoDB | - |
| `MONGODB_DB` | Nome do banco de dados | `auctions` |

Formatos aceitos para `AUCTION_DURATION`: `30s`, `5m`, `1h`, `2h30m`, etc.

### como rodar

#### com docker compose (recomendado)

```bash
docker-compose up --build
```

A aplicação estará disponível em `http://localhost:8080`.

#### local (requer MongoDB rodando)

```bash
go run cmd/auction/main.go
```

### endpoints

| Método | Rota | Descrição |
|---|---|---|
| POST | `/auction` | Criar leilão |
| GET | `/auction` | Listar leilões |
| GET | `/auction/:auctionId` | Buscar leilão por ID |
| GET | `/auction/winner/:auctionId` | Buscar lance vencedor |
| POST | `/bid` | Criar lance |
| GET | `/bid/:auctionId` | Listar lances por leilão |
| GET | `/user/:userId` | Buscar usuário |

### testes

#### teste de fechamento automático

O teste está em `internal/infra/database/auction/create_auction_test.go`.

requer MongoDB rodando (localhost:27017):

```bash
# subir apenas o MongoDB
docker-compose up mongodb -d

# rodar os testes
go test ./internal/infra/database/auction/ -v -run TestAuctionAutoClose -timeout 60s

# rodar teste unitário da função de duração (sem dependência de MongoDB)
go test ./internal/infra/database/auction/ -v -run TestGetAuctionDuration
```

O teste `TestAuctionAutoClose`:
1. Cria um leilão com `AUCTION_DURATION=3s`
2. Verifica que o status inicial é `Active`
3. Aguarda 4 segundos (duração + buffer)
4. Verifica que o status mudou para `Completed` automaticamente
