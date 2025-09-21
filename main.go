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

	// Inicia processador
	go processaJogo(&jogo)

	interfaceDesenharJogo(&jogo)

	for {
		evento := interfaceLerEventoTeclado()

		if !personagemExecutarAcao(evento, &jogo) {
			break
		}

		// A vida do personagem não mudava porque o símbolo do mapa era sobrescrito
		// antes de ser verificado. A correção é verificar o elemento que o personagem
		// acabou de pisar, que está armazenado em "UltimoVisitado".
		if jogo.UltimoVisitado.simbolo == 'x' {
			res := make(chan bool)
			canalJogo <- AcoesJogo{Acao: "dano", Valor: 3, Resposta: res}
			<-res
		}

		// Verifica se o jogador se moveu para um ponto de cura
		if jogo.UltimoVisitado.simbolo == '♥' {
			curar(&jogo)
		}

		interfaceDesenharJogo(&jogo)

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
