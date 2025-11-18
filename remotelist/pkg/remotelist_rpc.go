package remotelist

import (
	"bufio"        // Para ler o arquivo por linha de maneira eficaz
	"encoding/gob" // Para salvar e ler o map no arquivo
	"errors"
	"fmt"
	"io"      // Para posicionar o ponteiro do arquivo ao final ao escrever o log
	"log"     //Para gerar os Logs
	"os"      //Para lidar com os arquivos
	"strconv" // Converter o valor string para int
	"strings" // Para separar a linha após achar a virgula
	"sync"
	"time" // Para dar o corte de tempos em tempos
)

// Argumentos para o append
type AppendArgs struct {
	ListID string
	Value  int
}

// Argumentos para o Get
type GetArgs struct {
	ListID string
	Index  int // Posição 'i'
}

// Argumento responsavel pelo Remove e o Size
type ListArgs struct {
	ListID string
}

type RemoteList struct {
	mu      sync.Mutex
	list    map[string][]int //Transformado em Map para receber N listas
	logFile *os.File         // Faz a gravação dos logs no arquivo
}

func (l *RemoteList) loadSnapshot() {
	// Tenta abrir o arquivo de snapshot para Leitura
	file, err := os.Open("snapshot.dat")
	if err != nil {
		// Se não existe, é a primeira execução.
		if os.IsNotExist(err) {
			fmt.Println("Nenhum snapshot encontrado. Iniciando do zero.")
			return
		}
		log.Fatalf("Erro fatal ao abrir snapshot: %v", err)
	}
	defer file.Close()

	//Responsável pela decodificação, o conteudo indo direto para o l.list
	//&l.list é o ponteiro que chama a função que modifica o map
	fmt.Println("Carregando estado do snapshot...")
	if err := gob.NewDecoder(file).Decode(&l.list); err != nil {
		log.Fatalf("Erro fatal ao decodificar snapshot: %v", err)
	}
	fmt.Println("Snapshot carregado com sucesso.")
}

// Função para salvar o snapshot
func (l *RemoteList) saveSnapshot() {
	//Trava o mutex para garantir que nenhum cliente mude o map enquanto o snapshot é salvo
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Println("Iniciando snapshot...")

	//Abre o arquivo do snapshot para escrever
	file, err := os.Create("snapshot.dat")
	if err != nil {
		fmt.Printf("Erro ao criar snapshot: %v\n", err)
		return
	}
	defer file.Close()

	// Codifica o map inteiro para salvar no arquivo
	if err := gob.NewEncoder(file).Encode(l.list); err != nil {
		fmt.Printf("Erro ao codificar snapshot: %v\n", err)
		return
	}

	// Executa a limpeza do arquivo de log, após guardar ações após o snapshot.
	if err := l.logFile.Truncate(0); err != nil {
		fmt.Printf("Erro ao truncar log: %v\n", err)
	}

	//Move o cursor de volta para o inicio
	if _, err := l.logFile.Seek(0, 0); err != nil {
		fmt.Printf("Erro ao 'seek' log: %v\n", err)
	}

	fmt.Println("Snapshot salvo e log truncado com sucesso.")
}

// Snappshotter é a goroutine que roda em background ou segundo plano
func (l *RemoteList) snapshotter() {
	// Cria um timer que dispara a cada 30 segundos
	ticker := time.NewTicker(30 * time.Second)

	//Esse for ranger é um loop infinito que espera o timer disparar
	for range ticker.C {
		l.saveSnapshot()
	}
}

func (l *RemoteList) Append(args *AppendArgs, reply *bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Pega o ID da Lista que vai ser usado
	idDaLista := args.ListID
	// Pega a Lista selecionada pelo ID, se existir.
	lista, ok := l.list[idDaLista]
	// Caso a lista não exista, crie uma nova vazia
	if !ok {
		lista = make([]int, 0)
	}
	// Modifica a lista, com um novo valor adicionado
	lista = append(lista, args.Value)
	// Guarda a lista para dentro do Map
	l.list[idDaLista] = lista
	// Move o cursor para o final, simulando o modo append manualmente.
	if _, err := l.logFile.Seek(0, io.SeekEnd); err != nil {
		return errors.New("falha crítica: não foi possível posicionar o log")
	}
	//Log Recebe os valores na seguinte ordem: OPERAÇÃO, ID_DA_LISTA, VALOR
	logString := fmt.Sprintf("APPEND,%s,%d\n", args.ListID, args.Value)
	//Em caso de erro critico que não consiga salvar na escrita do log
	if _, err := l.logFile.WriteString(logString); err != nil {
		return errors.New("falha critica: Não foi possível salvar o log")
	}

	fmt.Println("Map atual: ", l.list) //Imprime o Map
	*reply = true
	return nil
}

func (l *RemoteList) Remove(args *ListArgs, reply *int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	//Pega o ID da Lista
	idDaLista := args.ListID
	//Se encontrado, Pega a Lista
	lista, ok := l.list[idDaLista]
	//Se não
	if !ok {
		return errors.New("lista não encontrada")
	}
	//Se a lista estiver vazia
	if len(lista) == 0 {
		return errors.New("lista vazia")
	}
	//Obtém o indice de ultima posição da lista
	ultimoIndice := len(lista) - 1
	//Guardando o valor que foi removido para retornar ao cliente
	*reply = lista[ultimoIndice]
	//Cria a lista com a remoção do ultimo elemento da lista anterior
	lista = lista[:ultimoIndice]
	//Atualiza a lista antiga para ter a nova lista com o elemento removido.
	l.list[idDaLista] = lista
	// Posiciona o cursor no final antes de registrar a operação.
	if _, err := l.logFile.Seek(0, io.SeekEnd); err != nil {
		return errors.New("falha crítica: não foi possível posicionar o log")
	}
	//Log Recebe os valores na seguinte ordem: OPERAÇÃO, ID_DA_LISTA, VALOR
	logString := fmt.Sprintf("REMOVE,%s\n", args.ListID)
	//Em caso de erro critico que não consiga salvar na escrita do log
	if _, err := l.logFile.WriteString(logString); err != nil {
		return errors.New("falha crítica: não foi possivel salvar log")
	}
	fmt.Println("Estado Atual do mapa: ", l.list)
	return nil
}

//Função que retorna o tamanho da lista

func (l *RemoteList) Size(args *ListArgs, reply *int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	//Pega o ID da Lista
	idDaLista := args.ListID
	//Se encontrado, Pega a Lista
	lista, ok := l.list[idDaLista]
	//Se não
	if !ok {
		return errors.New("lista não encontrada")
	}
	//Recebe o tamanho da lista
	*reply = len(lista)

	return nil //Em caso de sucesso
}

// Função para obter os elementos da lista
func (l *RemoteList) Get(args *GetArgs, reply *int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	//Pega o ID da lista
	idDaLista := args.ListID
	//Pega o Indice
	indice := args.Index

	//Pega a Lista do mapa caso OK
	lista, ok := l.list[idDaLista]
	//Se não
	if !ok {
		return errors.New("lista não encontrada")
	}
	//Se o indice for negativo ou estiver fora do tamanho da lista
	if indice < 0 || indice >= len(lista) {
		return errors.New("índice fora do intervalo da lista")
	}

	//O valor do elemento que está na posição indice da lista
	*reply = lista[indice]

	return nil //Em caso de sucesso

}

func NewRemoteList() *RemoteList {
	//Struct Criada
	r := new(RemoteList)

	// Tenta carregar o snapshot no primeiro momento
	r.loadSnapshot()

	//Caso o snapshot não exista, é inicializado com o make
	if r.list == nil {
		r.list = make(map[string][]int)
		fmt.Println("Mapa inicializado (snapshot era nil ou não existia).")
	}

	// Reconstroi o estado do r.list lendo o arquivo
	r.rebuildStateFromLog()

	//Abre o Arquivo
	// O_APPEND escreve sempre no final do arquivo
	// O_CREATE cria o arquivo, caso não exista
	// O_WRONLY Abre o Arquivo para a escrita
	// 0666 permissão padrão
	// Abre o log em modo leitura/escrita para permitir truncar durante o snapshot.
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Erro fatal ao abrir arquivo de log: %v", err)
	}
	// Garante que novas escritas sempre aconteçam ao final do arquivo.
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		log.Fatalf("Erro fatal ao posicionar log no final: %v", err)
	}
	r.logFile = file

	//Execução da goroutine que vai salvar o snapshot
	go r.snapshotter() //Go inicia uma nova thread

	return r
}

// Função responsável por ler o arquivo de log e repopular o map
func (l *RemoteList) rebuildStateFromLog() {
	//Tenta abrir o arquivo log para a leitura
	file, err := os.Open("log.txt")

	//Caso o arquivo não existe, é primeira inicialização.
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Arquivo de log não encontrado. Iniciando com estado vazio")
			return //Retorna, pois não há estado para recuperar
		}
		//Alerta para outros erros que não sejam o de arquivo não encontrado
		log.Fatalf("Erro fatal ao ler arquivo de log: %v", err)
	}
	//Garante que o arquivo será fechado ao final da função
	defer file.Close()

	fmt.Println("Recuperando estado do log...")

	// Executa a leitura de linha por linha do arquivo
	scanner := bufio.NewScanner(file)

	// Laço de repetição que lê cada linha
	for scanner.Scan() {
		line := scanner.Text() // Pega a linha

		// Quebra linha pela vírgula
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			fmt.Printf("Aviso: ignorando linha de log mal formada: %s\n", line)
			continue // Pula para a próxima linha
		}

		// Pega o tipo de operação e o ID da lista
		opType := parts[0]
		listID := parts[1]

		// Usa o switchpara dedicir qual passo adotar
		switch opType {
		case "APPEND":
			if len(parts) < 3 {
				fmt.Printf("Aviso: ignorando log 'APPEND' mal formado: %s\n", line)
				continue
			}

			// Converte o valor (parts[2]) de string para int
			value, err := strconv.Atoi(parts[2])
			if err != nil {
				fmt.Printf("Aviso: ignorando log 'APPEND' com valor inválido: %s\n", line)
				continue
			}

			// Fazendo a logica de append para o map.
			lista, ok := l.list[listID]
			if !ok {
				lista = make([]int, 0)
			}
			lista = append(lista, value)
			l.list[listID] = lista

		case "REMOVE":

			// Fazendo a lógica de remoção diretamente.
			lista, ok := l.list[listID]
			if ok && len(lista) > 0 {
				ultimoIndice := len(lista) - 1
				lista = lista[:ultimoIndice]
				l.list[listID] = lista
			} else {
				// Isso pode aocntecer se o log tiver corrompido,
				// mas não é um erro fatal, é apenas um aviso.
				fmt.Printf("Aviso: ignorando o log 'REMOVE' em lista vazia ou inexistente: %s\n", line)
			}
		}
	}
	// Verifica se o scanner parou por um erro (ex: arquivo corrompido)
	if err := scanner.Err(); err != nil {
		log.Fatalf("Erro fatal ao escanear arquivo de log: %v", err)
	}
	fmt.Println("Estado recuperado com sucesso. Estaod atual:", l.list)
}
