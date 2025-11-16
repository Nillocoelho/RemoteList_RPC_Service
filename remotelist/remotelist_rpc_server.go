package main

import (
	"log"
	"net"
	"net/rpc"

	remotelist "ifpb/remotelist/pkg"
)

func main() {
	// Usa o construtor para garantir que o mapa, o log e o snapshot sejam inicializados.
	list := remotelist.NewRemoteList()

	rpcs := rpc.NewServer()
	// Registra o serviço RPC e encerra caso haja falha crítica.
	if err := rpcs.Register(list); err != nil {
		log.Fatalf("falha ao registrar serviço RPC: %v", err)
	}

	// Escuta novas conexões na porta 5000.
	listener, err := net.Listen("tcp", "[localhost]:5000")
	if err != nil {
		log.Fatalf("erro ao iniciar listener: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// Mantém o servidor vivo, mas registra o erro para depuração.
			log.Printf("erro ao aceitar conexão: %v", err)
			continue
		}
		// Atende o cliente em uma goroutine para não bloquear novas conexões.
		go rpcs.ServeConn(conn)
	}
}
