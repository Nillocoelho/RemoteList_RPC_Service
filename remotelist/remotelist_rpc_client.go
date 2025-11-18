package main

import (
	"fmt"
	"log"
	"net/rpc"
	"ifpb/remotelist/pkg"
)

func main() {
	// Conecta ao servidor RPC
	client, err := rpc.Dial("tcp", "localhost:5000")
	if err != nil {
		log.Fatalf("Erro ao conectar ao servidor: %v", err)
	}
	defer client.Close()

	fmt.Println("Cliente conectado!")
	fmt.Println("--- Iniciando Testes ---")

	// --- 1. Testando Append ---
	var replyBool bool
	fmt.Println("Adicionando 10 e 20 na 'lista_A'...")

	argsAppendA1 := &remotelist.AppendArgs{ListID: "lista_A", Value: 10}
	err = client.Call("RemoteList.Append", argsAppendA1, &replyBool)
	if err != nil {
		log.Printf("Erro no Append A1: %v", err)
	}

	argsAppendA2 := &remotelist.AppendArgs{ListID: "lista_A", Value: 20}
	err = client.Call("RemoteList.Append", argsAppendA2, &replyBool)
	if err != nil {
		log.Printf("Erro no Append A2: %v", err)
	}

	fmt.Println("Adicionando 99 na 'lista_B'...")
	argsAppendB1 := &remotelist.AppendArgs{ListID: "lista_B", Value: 99}
	err = client.Call("RemoteList.Append", argsAppendB1, &replyBool)
	if err != nil {
		log.Printf("Erro no Append B1: %v", err)
	}

	// --- 2. Testando Size ---
	var replyInt int
	argsSizeA := &remotelist.ListArgs{ListID: "lista_A"}
	err = client.Call("RemoteList.Size", argsSizeA, &replyInt)
	if err != nil {
		log.Printf("Erro no Size (A): %v", err)
	} else {
		// Esperado: 2
		fmt.Printf("Tamanho da 'lista_A': %d\n", replyInt)
	}

	argsSizeB := &remotelist.ListArgs{ListID: "lista_B"}
	err = client.Call("RemoteList.Size", argsSizeB, &replyInt)
	if err != nil {
		log.Printf("Erro no Size (B): %v", err)
	} else {
		// Esperado: 1
		fmt.Printf("Tamanho da 'lista_B': %d\n", replyInt)
	}

	// --- 3. Testando Get ---
	// Pegando o item no índice 1 da lista_A (deve ser 20)
	argsGet := &remotelist.GetArgs{ListID: "lista_A", Index: 1}
	err = client.Call("RemoteList.Get", argsGet, &replyInt)
	if err != nil {
		log.Printf("Erro no Get (A,1): %v", err)
	} else {
		// Esperado: 20
		fmt.Printf("Valor no índice 1 da 'lista_A': %d\n", replyInt)
	}

	// Testando um erro (índice fora do limite)
	argsGetErr := &remotelist.GetArgs{ListID: "lista_A", Index: 99}
	err = client.Call("RemoteList.Get", argsGetErr, &replyInt)
	if err != nil {
		// Esperado: "índice fora do intervalo da lista"
		fmt.Printf("Erro esperado no Get (A,99): %v\n", err)
	}

	// --- 4. Testando Remove ---
	// Removendo o último item da lista_A (deve ser 20)
	argsRemoveA := &remotelist.ListArgs{ListID: "lista_A"}
	err = client.Call("RemoteList.Remove", argsRemoveA, &replyInt)
	if err != nil {
		log.Printf("Erro no Remove (A): %v", err)
	} else {
		// Esperado: 20
		fmt.Printf("Valor removido da 'lista_A': %d\n", replyInt)
	}

	// Checando o tamanho da lista_A de novo
	err = client.Call("RemoteList.Size", argsSizeA, &replyInt)
	if err != nil {
		log.Printf("Erro no Size (A) pós-Remove: %v", err)
	} else {
		// Esperado: 1
		fmt.Printf("Novo tamanho da 'lista_A': %d\n", replyInt)
	}

	fmt.Println("--- Testes Finalizados ---")
}
