// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool // Indica se o elemento bloqueia passagem
}

// Ação que são feitas sobre o jogador ou Inimigo
type AcoesJogo struct {
	X, Y     int
	Acao     string // "dano", "cura"
	Valor    int    // quantidade de HP a aplicar
	Resposta chan bool
}

type Boneco struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa           [][]Elemento // grade 2D representando o mapa
	PosX, PosY     int          // posição atual do personagem
	UltimoVisitado Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg      string       // mensagem para a barra de status
	Vida           int
	ChaveCapturada bool

	mu sync.RWMutex
}

var bausAbertos = 0

// Canal de comunicação
var canalJogo = make(chan AcoesJogo)

// Elementos visuais do jogo
var (
	Personagem = Boneco{'☺', termbox.ColorCyan, CorPadrao, true}
	Inimigo    = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
	Armadilha  = Elemento{'X', CorVermelho, CorPadrao, false}
	Cura       = Elemento{'♥', CorVerde, CorPadrao, false}
	Bau        = Elemento{'⌺', CorBau, CorPadrao, true}
	BauAberto  = Elemento{'⍓', CorVermelho, CorPadrao, true}
	Alcapao    = Elemento{'⍋', CorAzul, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	return Jogo{UltimoVisitado: Vazio, Vida: 10, ChaveCapturada: false}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()

	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				e = Inimigo
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y // registra a posição inicial do personagem
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Semeia o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Adiciona as armadilhas e pontos de cura em posições aleatórias
	for i := 0; i < 6; i++ {
		posicionarAleatoriamente(jogo, Armadilha)
		posicionarAleatoriamente(jogo, Bau)
	}
	for i := 0; i < 3; i++ {
		posicionarAleatoriamente(jogo, Cura)
	}

	posicionarAleatoriamente(jogo, Alcapao)
	return nil
}

// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	// Verifica se a coordenada Y está dentro dos limites verticais do mapa
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}

	// Verifica se a coordenada X está dentro dos limites horizontais do mapa
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	// Verifica se o elemento de destino é tangível (bloqueia passagem)
	if jogo.Mapa[y][x].tangivel {
		return false
	}

	// Pode mover para a posição
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy

	// Obtem elemento atual na posição
	elemento := jogo.Mapa[y][x] // guarda o conteúdo atual da posição

	jogo.Mapa[y][x] = jogo.UltimoVisitado   // restaura o conteúdo anterior
	jogo.UltimoVisitado = jogo.Mapa[ny][nx] // guarda o conteúdo atual da nova posição
	jogo.Mapa[ny][nx] = elemento            // move o elemento
}

// Processa as requisições recebidas pelo canalJogo e atualiza o estado do jogo
func processaJogo(jogo *Jogo) {
	for req := range canalJogo {
		jogo.mu.Lock()
		switch req.Acao {
		case "dano":
			jogo.Vida -= req.Valor
			if jogo.Vida < 0 {
				jogo.Vida = 0
			}
		case "cura":
			jogo.Vida += req.Valor
			if jogo.Vida > 10 {
				jogo.Vida = 10
			}
		}
		jogo.mu.Unlock()
		req.Resposta <- true
	}
}

func curar(jogo *Jogo) {
	curaX, curaY := jogo.PosX, jogo.PosY

	go func() {
		jogo.StatusMsg = "Você está em um ponto de cura! Aguarde 3s para comecar a curar..."
		interfaceDesenharJogo(jogo)

		time.Sleep(3 * time.Second)

		for {
			jogo.mu.RLock()
			vidaAtual := jogo.Vida
			posAtualX, posAtualY := jogo.PosX, jogo.PosY
			jogo.mu.RUnlock()

			if (posAtualX != curaX || posAtualY != curaY) || vidaAtual >= 10 {
				if vidaAtual >= 10 {
					jogo.StatusMsg = "Vida totalmente restaurada!"
				} else {
					jogo.StatusMsg = "Você saiu do ponto de cura."
				}
				interfaceDesenharJogo(jogo)
				return
			}

			// Envia o pedido de cura.
			res := make(chan bool)
			canalJogo <- AcoesJogo{Acao: "cura", Valor: 1, Resposta: res}
			<-res

			jogo.mu.RLock()
			vidaAtual = jogo.Vida
			jogo.mu.RUnlock()
			jogo.StatusMsg = "Curando... Vida: " + strconv.Itoa(vidaAtual)
			interfaceDesenharJogo(jogo)

			time.Sleep(1 * time.Second)
		}
	}()
}

func gameOver(jogo *Jogo) {
	jogo.mu.Lock()
	if jogo.ChaveCapturada {
		jogo.StatusMsg = "Voce GANHOU! PARABENS!!!!! Pressione ESC para sair."
	} else if jogo.Vida <= 0 {
		jogo.StatusMsg = "Voce PERDEU! Mais sorte na proxima! Pressione ESC para sair."
	} else {
		jogo.StatusMsg = "GAME OVER! Pressione ESC para sair."
	}
	jogo.mu.Unlock()
	interfaceDesenharJogo(jogo)

	for {
		ev := interfaceLerEventoTeclado()
		if ev.Tipo == "sair" {
			return
		}
	}
}

func posicionarAleatoriamente(jogo *Jogo, elemento Elemento) {
	for {
		y := rand.Intn(len(jogo.Mapa))
		x := rand.Intn(len(jogo.Mapa[0]))

		// Verifica se a posição está vazia
		if jogo.Mapa[y][x].simbolo == Vazio.simbolo {
			jogo.Mapa[y][x] = elemento
			return
		}
	}
}

func podeSair(jogo *Jogo) {
	if jogo.ChaveCapturada {
		jogo.StatusMsg = "Parabens, voce conseguiu sair!"
		time.Sleep(3 * time.Second)
		gameOver(jogo)
	} else {
		jogo.StatusMsg = "Voce nao pode sair, voce nao encontrou a Cave"
	}

}

func abrirBau(jogo *Jogo, x, y int) {
	jogo.mu.RLock()
	defer jogo.mu.RUnlock()

	jogo.Mapa[y][x] = BauAberto
	bausAbertos += 1

	if ((rand.Intn(100) < 30) || (bausAbertos == 6)) && !jogo.ChaveCapturada {
		jogo.StatusMsg = "CHAVE ENCONTRADA!"
		jogo.ChaveCapturada = true
	} else {
		switch rand.Intn(2) {
		case 0:
			jogo.StatusMsg = "O bau esta vazio"
			time.Sleep(3 * time.Second)
		case 1:
			jogo.Vida -= 1
			jogo.StatusMsg = "Tinha uma armadilha no Bau"
			time.Sleep(3 * time.Second)
			if jogo.Vida <= 0 {
				gameOver(jogo)
			}

		}
	}
}
