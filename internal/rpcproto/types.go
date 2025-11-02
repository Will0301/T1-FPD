package rpcproto

import "time"

type RegistrarArgs struct {
	Nome string
}

type RegistrarResp struct {
	Mensagem string
}

type AtualizarArgs struct {
	Nome  string
	X, Y  int
	Vida  int
	Chave bool
	Seq   uint64
}

type AtualizarResp struct {
	Mensagem string
	Seq      uint64
}

type EstadoArgs struct{}

type Jogador struct {
	Nome   string
	X, Y   int
	Vida   int
	Chave  bool
	Ultimo time.Time
}

type EstadoResp struct {
	Jogadores []Jogador
}
