package main

import (
	"sync"
	"time"

	"jogo/internal/rpcproto"
)

type (
	RegistrarArgs = rpcproto.RegistrarArgs
	RegistrarResp = rpcproto.RegistrarResp
	AtualizarArgs = rpcproto.AtualizarArgs
	AtualizarResp = rpcproto.AtualizarResp
	EstadoArgs    = rpcproto.EstadoArgs
	Jogador       = rpcproto.Jogador
	EstadoResp    = rpcproto.EstadoResp
)

type JogadorRemoto struct {
	Nome   string
	X, Y   int
	Vida   int
	Chave  bool
	Ultimo time.Time
}

var jogadoresRemotos struct {
	sync.RWMutex
	m map[string]JogadorRemoto
}
