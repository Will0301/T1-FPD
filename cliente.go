package main

import (
	"log"
	"net/rpc"
	"time"
)

type ClienteRPC struct {
	conexao *rpc.Client
	Nome    string
	seq     uint64
}

func NovoClienteRPC(endereco, nome string) (*ClienteRPC, error) {
	c, err := rpc.Dial("tcp", endereco)
	if err != nil {
		return nil, err
	}
	cli := &ClienteRPC{conexao: c, Nome: nome}
	var resp RegistrarResp
	err = c.Call("Servidor.Registrar", RegistrarArgs{Nome: nome}, &resp)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (c *ClienteRPC) EnviarAtualizacao(x, y, vida int) {
	var resp AtualizarResp
	c.seq++
	args := AtualizarArgs{
		Nome: c.Nome,
		X:    x,
		Y:    y,
		Vida: vida,
		Seq:  c.seq,
	}

	err := retryRPC(func() error {
		return c.conexao.Call("Servidor.Atualizar", args, &resp)
	})
	if err != nil {
		log.Printf("erro ao enviar atualizacao: %v", err)
	}
}

func retryRPC(fn func() error) error {
	const tentativas = 3
	for i := 1; i <= tentativas; i++ {
		if err := fn(); err != nil {
			if i == tentativas {
				return err
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}
		return nil
	}
	return nil
}

func (c *ClienteRPC) AtualizarEstado() {
	go func() {
		for {
			var resp EstadoResp
			err := retryRPC(func() error {
				return c.conexao.Call("Servidor.ObterEstado", EstadoArgs{}, &resp)
			})
			if err != nil {
				log.Printf("erro ao obter estado: %v", err)
				time.Sleep(time.Second)
				continue
			}

			jogadoresRemotos.Lock()
			if jogadoresRemotos.m == nil {
				jogadoresRemotos.m = make(map[string]JogadorRemoto)
			}
			for _, j := range resp.Jogadores {
				if j.Nome == c.Nome {
					continue
				}
				jogadoresRemotos.m[j.Nome] = JogadorRemoto{
					Nome:   j.Nome,
					X:      j.X,
					Y:      j.Y,
					Vida:   j.Vida,
					Ultimo: j.Ultimo,
				}
			}
			for nome, j := range jogadoresRemotos.m {
				if time.Since(j.Ultimo) > 10*time.Second {
					delete(jogadoresRemotos.m, nome)
				}
			}
			jogadoresRemotos.Unlock()

			solicitarRedesenho()
			time.Sleep(500 * time.Millisecond)
		}
	}()
}
