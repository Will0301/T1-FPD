// personagem.go - Funções para movimentação e ações do personagem
package main

import "fmt"

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1 // Move para cima
	case 'a':
		dx = -1 // Move para a esquerda
	case 's':
		dy = 1 // Move para baixo
	case 'd':
		dx = 1 // Move para a direita
	default:
		return // Não é uma tecla de movimento, não faz nada
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy

	resposta := make(chan bool)
	acao := AcaoMapa{
		Tipo:     TipoMoverJogador,
		DestinoX: nx,
		DestinoY: ny,
		Resposta: resposta,
	}

	// Envia a solicitação de movimento para o gerenciador do mapa
	jogo.CanalMapa <- acao

	// Aguarda a resposta. A posição do personagem (jogo.PosX/Y)
	// será atualizada dentro da goroutine 'processaMapa' se o movimento for válido.
	<-resposta
}

// Define o que ocorre quando o jogador pressiona a tecla de interação
// Neste exemplo, apenas exibe uma mensagem de status
// Você pode expandir essa função para incluir lógica de interação com objetos
func personagemInteragir(jogo *Jogo) {
	jogo.mu.RLock()
	px, py := jogo.PosX, jogo.PosY
	jogo.mu.RUnlock()

	posicoesParaVerificar := [][2]int{
		{px, py - 1}, // Cima
		{px, py + 1}, // Baixo
		{px - 1, py}, // Esquerda
		{px + 1, py}, // Direita
	}
	for _, pos := range posicoesParaVerificar {
		x, y := pos[0], pos[1]

		if y >= 0 && y < len(jogo.Mapa) && x >= 0 && x < len(jogo.Mapa[y]) {
			jogo.mu.RLock()
			simboloNoLocal := jogo.Mapa[y][x].simbolo
			jogo.mu.RUnlock()

			if simboloNoLocal == Bau.simbolo {
				jogo.StatusMsg = fmt.Sprintf("Abrindo baú em (%d, %d)...", x, y)
				abrirBau(jogo, x, y)
				return
			}
			if simboloNoLocal == BauAberto.simbolo {
				jogo.StatusMsg = "Este baú já foi aberto."
			}
		}
	}
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		return false // Indica que o jogo deve terminar
	case "interagir":
		// Executa a ação de interação
		personagemInteragir(jogo)
	case "mover":
		// Move o personagem com base na tecla
		personagemMover(ev.Tecla, jogo)
	}
	return true // Continua o jogo
}
