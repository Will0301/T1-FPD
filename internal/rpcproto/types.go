package rpcproto

import "time"

type RegistrarArgs struct {
	Nome string
}

type RegistrarResp struct {
	Mensagem string
}

type AtualizarArgs struct {
	Nome string
	X, Y int
	Vida int
}

type AtualizarResp struct {
	Mensagem string
}

type EstadoArgs struct{}

type Jogador struct {
	Nome   string
	X, Y   int
	Vida   int
	Ultimo time.Time
}

type EstadoResp struct {
	Jogadores []Jogador
}
