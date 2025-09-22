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

		// A verificação de dano e cura funciona de forma segura, pois a ação de mover
		// no canal bloqueia até que 'jogo.UltimoVisitado' seja atualizado.
		if jogo.UltimoVisitado.simbolo == 'x' {
			res := make(chan bool)
			canalJogo <- AcoesJogo{Acao: "dano", Valor: 3, Resposta: res}
			<-res
			// --- CORRIGIDO: Solicita um redesenho de forma segura ---
			select {
			case jogo.CanalRedesenhar <- true:
			default: // Não bloqueia se já houver uma solicitação pendente
			}
		}

		if jogo.UltimoVisitado.simbolo == '♥' {
			curar(&jogo)
		}

		// A chamada de interfaceDesenharJogo() foi removida daqui,
		// pois o gerenciador 'processaMapa' agora centraliza todos os redesenhos.

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
