package main

import (
	"fmt"
	"net/rpc"
	"time"
)

type ClienteRPC struct {
	conexao *rpc.Client
	Nome    string
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
	fmt.Println(resp.Mensagem)
	return cli, nil
}

func (c *ClienteRPC) EnviarAtualizacao(x, y, vida int) {
	var resp AtualizarResp
	args := AtualizarArgs{Nome: c.Nome, X: x, Y: y, Vida: vida}
	err := c.conexao.Call("Servidor.Atualizar", args, &resp)
	if err == nil {
		fmt.Println(resp.Mensagem)
	}
}

func (c *ClienteRPC) AtualizarEstado() {
	go func() {
		for {
			var resp EstadoResp
			err := c.conexao.Call("Servidor.ObterEstado", EstadoArgs{}, &resp)
			if err == nil {
				fmt.Println("Jogadores conectados:")
				for _, j := range resp.Jogadores {
					fmt.Printf(" - %s (%d,%d) vida=%d\n", j.Nome, j.X, j.Y, j.Vida)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}()
}
