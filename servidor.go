package main

import (
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Servidor struct {
	mu       sync.Mutex
	jogadores map[string]*Jogador
}

func NovoServidor() *Servidor {
	return &Servidor{jogadores: make(map[string]*Jogador)}
}

func (s *Servidor) Registrar(args RegistrarArgs, resp *RegistrarResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.jogadores[args.Nome]; !ok {
		s.jogadores[args.Nome] = &Jogador{Nome: args.Nome, X: 0, Y: 0, Vida: 10, Ultimo: time.Now()}
		log.Printf("Jogador conectado: %s\n", args.Nome)
	}
	resp.Mensagem = "Registrado"
	return nil
}

func (s *Servidor) Atualizar(args AtualizarArgs, resp *AtualizarResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jogadores[args.Nome]; ok {
		j.X = args.X
		j.Y = args.Y
		j.Vida = args.Vida
		j.Ultimo = time.Now()
		log.Printf("%s -> (%d,%d) vida=%d\n", j.Nome, j.X, j.Y, j.Vida)
	} else {
		resp.Mensagem = "Jogador não registrado"
	}
	return nil
}

func (s *Servidor) ObterEstado(args EstadoArgs, resp *EstadoResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, j := range s.jogadores {
		resp.Jogadores = append(resp.Jogadores, *j)
	}
	return nil
}

func main() {
	s := NovoServidor()
    rpc.Register(s)

    ouvinte, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Servidor RPC ativo na porta 8080")
    rpc.Accept(ouvinte)
}
