package main

import (
	"os"
)

func main() {
	interfaceIniciar()
	defer interfaceFinalizar()

	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// Inicia os processadores concorrentes do jogo
	go processaJogo(&jogo)
	go processaMapa(&jogo)

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
			default: 
			}
		}

		if jogo.UltimoVisitado.simbolo == Cura.simbolo {
			curar(&jogo)
		}

		if jogo.UltimoVisitado.simbolo == Alcapao.simbolo {
			podeSair(&jogo)
		}

		interfaceDesenharJogo(&jogo)

		// Verifica se a vida chegou a 0 para encerrar o jogo
		jogo.mu.RLock()
		if jogo.Vida <= 0 {
			jogo.mu.RUnlock()
			gameOver(&jogo)
			return // Encerra a função main e o programa
		}
		jogo.mu.RUnlock()
	}
}
