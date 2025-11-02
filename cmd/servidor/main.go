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
	mu         sync.Mutex
	jogadores  map[string]*rpcproto.Jogador
	sequencias map[string]uint64
}

func NovoServidor() *Servidor {
	return &Servidor{
		jogadores:  make(map[string]*rpcproto.Jogador),
		sequencias: make(map[string]uint64),
	}
}

func (s *Servidor) Registrar(args rpcproto.RegistrarArgs, resp *rpcproto.RegistrarResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.jogadores[args.Nome]; !ok {
		s.jogadores[args.Nome] = &rpcproto.Jogador{
			Nome:   args.Nome,
			Ultimo: time.Now(),
		}
		s.sequencias[args.Nome] = 0
	}
	resp.Mensagem = "Registrado"
	log.Printf("Registrar <= %+v", args)
	return nil
}

func (s *Servidor) Atualizar(args rpcproto.AtualizarArgs, resp *rpcproto.AtualizarResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if j, ok := s.jogadores[args.Nome]; ok {
		ultimoSeq := s.sequencias[args.Nome]
		if args.Seq <= ultimoSeq {
			resp.Mensagem = "Duplicado"
			resp.Seq = ultimoSeq
			log.Printf("Atualizar DUP nome=%s seq=%d ignorado (ultimo=%d)", args.Nome, args.Seq, ultimoSeq)
		} else {
			j.X = args.X
			j.Y = args.Y
			j.Vida = args.Vida
			j.Ultimo = time.Now()
			s.sequencias[args.Nome] = args.Seq
			resp.Mensagem = "Atualizado"
			resp.Seq = args.Seq
			log.Printf("Atualizar OK %+v", args)
		}
	} else {
		resp.Mensagem = "Jogador não registrado"
		log.Printf("Atualizar FAIL nome=%s nao registrado", args.Nome)
	}

	log.Printf("Atualizar => %+v", resp)
	return nil
}

func (s *Servidor) ObterEstado(args rpcproto.EstadoArgs, resp *rpcproto.EstadoResp) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	agora := time.Now()
	for nome, j := range s.jogadores {
		if agora.Sub(j.Ultimo) > 10*time.Second {
			delete(s.jogadores, nome)
			delete(s.sequencias, nome)
			log.Printf("Removendo jogador inativo: %s", nome)
		}
	}

	resp.Jogadores = resp.Jogadores[:0]
	for _, j := range s.jogadores {
		resp.Jogadores = append(resp.Jogadores, *j)
	}

	log.Printf("ObterEstado => %d jogadores", len(resp.Jogadores))
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
