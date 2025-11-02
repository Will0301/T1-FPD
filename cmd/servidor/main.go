package main

import (
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"jogo/internal/rpcproto"
)

type Servidor struct {
	mu        sync.Mutex
	jogadores map[string]*rpcproto.Jogador
}

func NovoServidor() *Servidor {
	return &Servidor{jogadores: make(map[string]*rpcproto.Jogador)}
}

func (s *Servidor) Registrar(args rpcproto.RegistrarArgs, resp *rpcproto.RegistrarResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.jogadores[args.Nome]; !ok {
		s.jogadores[args.Nome] = &rpcproto.Jogador{Nome: args.Nome, X: 0, Y: 0, Vida: 10, Ultimo: time.Now()}
		log.Printf("Jogador conectado: %s\n", args.Nome)
	}

	resp.Mensagem = "Registrado"
	return nil
}

func (s *Servidor) Atualizar(args rpcproto.AtualizarArgs, resp *rpcproto.AtualizarResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if j, ok := s.jogadores[args.Nome]; ok {
		j.X = args.X
		j.Y = args.Y
		j.Vida = args.Vida
		j.Ultimo = time.Now()
		log.Printf("%s -> (%d,%d) vida=%d\n", j.Nome, j.X, j.Y, j.Vida)
		resp.Mensagem = "Atualizado"
	} else {
		resp.Mensagem = "Jogador não registrado"
	}

	return nil
}

func (s *Servidor) ObterEstado(args rpcproto.EstadoArgs, resp *rpcproto.EstadoResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp.Jogadores = resp.Jogadores[:0]
	for _, j := range s.jogadores {
		resp.Jogadores = append(resp.Jogadores, *j)
	}

	return nil
}

func main() {
	s := NovoServidor()
	if err := rpc.Register(s); err != nil {
		log.Fatalf("falha ao registrar servidor RPC: %v", err)
	}

	ouvinte, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ouvinte.Close()

	log.Println("Servidor RPC ativo na porta 8080")
	rpc.Accept(ouvinte)
}
