# RemoteList RPC Service

Aplicacao exemplo em Go que expoe uma lista remota via `net/rpc`. O servidor mantem diversas listas identificadas por string, suporta operacoes de append, leitura, remocao e tamanho, e persiste o estado em disco usando snapshot periodico e log incremental.

## Estrutura do Projeto

```
remotelist/
├── go.mod
├── remotelist_rpc_client.go   # Programa cliente CLI para exercitar o servico
├── remotelist_rpc_server.go   # Servidor RPC que publica o servico RemoteList
└── pkg/
    └── remotelist_rpc.go      # Implementacao da estrutura RemoteList e utilitarios de persistencia
```

## Requisitos

- Go 1.21 ou superior (`go env GOVERSION` para conferir)
- Sistemas testados: Windows 10/11 com PowerShell 5.1

## Como Executar

1. Abra dois terminais PowerShell na pasta `remotelist`.
2. No primeiro terminal, inicialize o servidor:
   ```powershell
   go run .\remotelist_rpc_server.go
   ```
   O servidor escuta em `localhost:5000` e cria/usa os arquivos `snapshot.dat` e `log.txt` para persistencia.
3. No segundo terminal, execute o cliente de teste que invoca todas as operacoes basicas:
   ```powershell
   go run .\remotelist_rpc_client.go
   ```
   O cliente imprime o resultado de cada chamada e avisa caso encontre algum erro esperado (por exemplo, acesso fora do intervalo).

## Operacoes Disponiveis

- `Append(listID string, value int)` adiciona um valor ao final da lista `listID`.
- `Remove(listID string)` remove e retorna o ultimo valor da lista.
- `Size(listID string)` retorna a quantidade de elementos da lista.
- `Get(listID string, index int)` obtem o elemento localizado na posicao `index`.

Todas as operacoes bloqueiam com `sync.Mutex` para garantir consistencia quando multiplos clientes efetuam chamadas simultaneas.

## Persistencia e Recuperacao

- A cada 30 segundos o servidor grava um snapshot completo em `snapshot.dat` e limpa o log de operacoes.
- Cada chamada mutavel (`Append` e `Remove`) eh registrada em `log.txt`. Em caso de reinicio, o servidor carrega o snapshot e reaplica as operacoes do log para reconstruir o estado.

## Logs

- Mensagens informativas sao impressas no terminal do servidor mostrando o mapa em memoria e advertencias durante a recuperacao.
- Logs de erro critico (por exemplo, falha ao abrir arquivos) encerram o processo com `log.Fatalf`.

## Dicas de Desenvolvimento

- Utilize `gofmt` para padronizar o codigo:
  ```powershell
  gofmt -w pkg\remotelist_rpc.go remotelist_rpc_server.go remotelist_rpc_client.go
  ```
- Para limpar arquivos de estado durante testes, remova `snapshot.dat` e `log.txt` antes de subir o servidor.

## Possiveis Extensoes

- Expor metodos de listagem completa ou reset do estado.
- Trocar `net/rpc` por gRPC ou HTTP JSON.
- Criar testes automatizados usando o pacote `testing` e um servidor RPC em memoria.
