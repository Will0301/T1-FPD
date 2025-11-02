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

type TipoAcao int

const (
	TipoMoverJogador TipoAcao = iota
	TipoMoverInimigo
)

type AcaoMapa struct {
	Tipo               TipoAcao
	OrigemX, OrigemY   int
	DestinoX, DestinoY int
	ElementoMovido     Elemento
	Resposta           chan bool
}

// Struct para enviar a posição do jogador aos inimigos
type PosicaoJogador struct {
	X, Y int
}

type Boneco struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa            [][]Elemento // grade 2D representando o mapa
	PosX, PosY      int          // posição atual do personagem
	UltimoVisitado  Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg       string       // mensagem para a barra de status
	Vida            int
	Inimigos        []*Inimigo
	CanalMapa       chan AcaoMapa
	CanalPosJogador chan PosicaoJogador // Canal para transmitir a posição do jogador
	CanalRedesenhar chan bool
	ChaveCapturada  bool

	mu sync.RWMutex
}

var bausAbertos = 0

// Canal de comunicação
var canalJogo = make(chan AcoesJogo)

// Elementos visuais do jogo
var (
	Personagem      = Boneco{'☺', termbox.ColorCyan, CorPadrao, true}
	InimigoElemento = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede          = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao       = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio           = Elemento{' ', CorPadrao, CorPadrao, false}
	Armadilha       = Elemento{'X', CorVermelho, CorPadrao, false}
	Cura            = Elemento{'♥', CorVerde, CorPadrao, false}
	Bau             = Elemento{'⌺', CorBau, CorPadrao, true}
	BauAberto       = Elemento{'⍓', CorVermelho, CorPadrao, true}
	Alcapao         = Elemento{'⍋', CorAzul, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	return Jogo{
		UltimoVisitado:  Vazio,
		Vida:            10,
		CanalMapa:       make(chan AcaoMapa),
		CanalPosJogador: make(chan PosicaoJogador, 10),
		CanalRedesenhar: make(chan bool, 1),
		ChaveCapturada:  false,
	}
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
			case InimigoElemento.simbolo:
				inimigo := novoInimigo(x, y, jogo.CanalMapa, canalJogo)
				jogo.Inimigos = append(jogo.Inimigos, inimigo)
				go inimigo.Executar(jogo.CanalPosJogador)
				e = inimigo.elemento
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Transmite a posição inicial do jogador para todos os inimigos
	initialPos := PosicaoJogador{X: jogo.PosX, Y: jogo.PosY}
	for _, inimigo := range jogo.Inimigos {
		inimigo.UltimoXJogador = initialPos.X
		inimigo.UltimoYJogador = initialPos.Y
	}

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
	if y < 0 || y >= len(jogo.Mapa) || x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	// Verifica se o elemento de destino é tangível (bloqueia passagem)
	if jogo.Mapa[y][x].tangivel {
		return false
	}

	return true
}

// goroutine que pode desenhar na tela
func processaMapa(jogo *Jogo) {
	for {
		select {
		case acao := <-jogo.CanalMapa:
			podeMover := false
			switch acao.Tipo {
			case TipoMoverJogador:
				podeMover = jogoPodeMoverPara(jogo, acao.DestinoX, acao.DestinoY)
				if podeMover {
					nx, ny := acao.DestinoX, acao.DestinoY
					jogo.Mapa[jogo.PosY][jogo.PosX] = jogo.UltimoVisitado
					jogo.UltimoVisitado = jogo.Mapa[ny][nx]
					jogo.Mapa[ny][nx] = Vazio
					jogo.PosX, jogo.PosY = nx, ny

					select {
					case jogo.CanalPosJogador <- PosicaoJogador{X: jogo.PosX, Y: jogo.PosY}:
					default:
					}
				}
			case TipoMoverInimigo:
				podeMover = jogoPodeMoverPara(jogo, acao.DestinoX, acao.DestinoY)
				if podeMover {
					jogo.Mapa[acao.DestinoY][acao.DestinoX] = acao.ElementoMovido
					jogo.Mapa[acao.OrigemY][acao.OrigemX] = Vazio
				}
			}
			acao.Resposta <- podeMover
			interfaceDesenharJogo(jogo) // Redesenha após uma ação no mapa
		case <-jogo.CanalRedesenhar:
			interfaceDesenharJogo(jogo) // Redesenha a pedido de outra goroutine
		}
	}
}

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

		// Solicita um redesenho de forma não-bloqueante
		select {
		case jogo.CanalRedesenhar <- true:
		default:
		}
	}
}

// Não desenha mais na tela, apenas solicita o redesenho
func curar(jogo *Jogo) {
	curaX, curaY := jogo.PosX, jogo.PosY

	go func() {
		jogo.StatusMsg = "Você está em um ponto de cura! Aguarde 3s para começar a curar..."
		select {
		case jogo.CanalRedesenhar <- true:
		default:
		}

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
				select {
				case jogo.CanalRedesenhar <- true:
				default:
				}
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
			select {
			case jogo.CanalRedesenhar <- true:
			default:
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func gameOver(jogo *Jogo) {
	jogo.mu.Lock()
	if jogo.ChaveCapturada && jogo.Vida > 0 {
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
		}
	}
}
