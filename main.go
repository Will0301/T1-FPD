package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	jogadoresRemotos.m = make(map[string]JogadorRemoto)

	interfaceIniciar()
	defer interfaceFinalizar()

	rand.Seed(time.Now().UnixNano())
	idJogador := fmt.Sprintf("Jogador-%d-%d", time.Now().UnixNano(), rand.Intn(1000))

	var err error
	clienteRPC, err = NovoClienteRPC("localhost:8080", idJogador)
	if err != nil {
		log.Fatal(err)
	}
	clienteRPC.AtualizarEstado()

	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	jogo := jogoNovo()
	jogoAtual = &jogo
	jogoCarregarMapa(mapaFile, &jogo)

	if clienteRPC != nil {
		clienteRPC.EnviarAtualizacao(jogo.PosX, jogo.PosY, jogo.Vida)
	}

	// Inicia os processadores concorrentes do jogo
	go processaJogo(&jogo)
	go processaMapa(&jogo)

	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		for range t.C {
			jogo.mu.RLock()
			x, y, vida := jogo.PosX, jogo.PosY, jogo.Vida
			jogo.mu.RUnlock()
			if clienteRPC != nil {
				clienteRPC.EnviarAtualizacao(x, y, vida)
			}
		}
	}()

	// Desenha o estado inicial do jogo
	interfaceDesenharJogo(&jogo)

	for {
		evento := interfaceLerEventoTeclado()

		// Sai do loop se a ação retornar 'false' (pressionou ESC)
		if !personagemExecutarAcao(evento, &jogo) {
			break
		}

		if jogo.UltimoVisitado.simbolo == Armadilha.simbolo {
			res := make(chan bool)
			canalJogo <- AcoesJogo{Acao: "dano", Valor: 3, Resposta: res}
			<-res
			select {
			case jogo.CanalRedesenhar <- true:
			default: // Não bloqueia se já houver uma solicitação pendente
			}
		}

		// Verifica se o jogador se moveu para um ponto de cura
		if jogo.UltimoVisitado.simbolo == Cura.simbolo {
			curar(&jogo)
		}

		if jogo.UltimoVisitado.simbolo == Alcapao.simbolo {
			podeSair(&jogo)
		}

		// Verifica se a vida chegou a 0 para encerrar o jogo
		jogo.mu.RLock()
		if jogo.Vida <= 0 {
			jogo.mu.RUnlock()
			gameOver(&jogo)
			return
		}
		jogo.mu.RUnlock()
	}
}
