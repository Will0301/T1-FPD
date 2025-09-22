package main

import (
	"math"
	"math/rand"
	"time"
)

const (
	EstadoPatrulhando = "PATRULHANDO"
	EstadoPerseguindo = "PERSEGUINDO"
	DistanciaVisao    = 8
)

// Inimigo representa um inimigo autônomo no jogo.
type Inimigo struct {
	x, y int
	// --- CORRIGIDO: Campos exportados com letra maiúscula ---
	UltimoXJogador, UltimoYJogador int
	estado                         string
	elemento                       Elemento
	canalMapa                      chan<- AcaoMapa
	canalJogo                      chan<- AcoesJogo
	ticker                         *time.Ticker
}

// novoInimigo cria uma nova instância de Inimigo.
func novoInimigo(x, y int, canalMapa chan<- AcaoMapa, canalJogo chan<- AcoesJogo) *Inimigo {
	return &Inimigo{
		x:         x,
		y:         y,
		estado:    EstadoPatrulhando,
		elemento:  InimigoElemento,
		canalMapa: canalMapa,
		canalJogo: canalJogo,
		ticker:    time.NewTicker(1 * time.Second),
	}
}

// Executar é o loop principal do inimigo.
func (i *Inimigo) Executar(canalPosJogador <-chan PosicaoJogador) {
	for {
		select {
		case posJogador := <-canalPosJogador:
			i.UltimoXJogador, i.UltimoYJogador = posJogador.X, posJogador.Y
			dist := distancia(i.x, i.y, posJogador.X, posJogador.Y)

			if dist <= DistanciaVisao {
				i.estado = EstadoPerseguindo
			} else {
				i.estado = EstadoPatrulhando
			}

		case <-i.ticker.C:
			if i.estado == EstadoPerseguindo {
				i.perseguirOuAtacar()
			} else {
				i.moverAleatoriamente()
			}
		}
	}
}

func (i *Inimigo) perseguirOuAtacar() {
	if distancia(i.x, i.y, i.UltimoXJogador, i.UltimoYJogador) == 1 {
		i.atacar()
		return
	}
	i.moverEmDirecao(i.UltimoXJogador, i.UltimoYJogador)
}

func (i *Inimigo) atacar() {
	resposta := make(chan bool)
	i.canalJogo <- AcoesJogo{Acao: "dano", Valor: 1, Resposta: resposta}
	<-resposta
}

func (i *Inimigo) moverEmDirecao(alvoX, alvoY int) {
	dx, dy := 0, 0
	if alvoX > i.x {
		dx = 1
	} else if alvoX < i.x {
		dx = -1
	}

	if alvoY > i.y {
		dy = 1
	} else if alvoY < i.y {
		dy = -1
	}

	if dx != 0 && dy != 0 {
		if rand.Intn(2) == 0 {
			dx = 0
		} else {
			dy = 0
		}
	}

	i.tentarMover(dx, dy)
}

func (i *Inimigo) moverAleatoriamente() {
	direcao := rand.Intn(4)
	dx, dy := 0, 0
	switch direcao {
	case 0:
		dy = -1
	case 1:
		dy = 1
	case 2:
		dx = -1
	case 3:
		dx = 1
	}
	i.tentarMover(dx, dy)
}

func (i *Inimigo) tentarMover(dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}
	novoX, novoY := i.x+dx, i.y+dy

	resposta := make(chan bool)
	acao := AcaoMapa{
		Tipo:           TipoMoverInimigo,
		OrigemX:        i.x,
		OrigemY:        i.y,
		DestinoX:       novoX,
		DestinoY:       novoY,
		ElementoMovido: i.elemento,
		Resposta:       resposta,
	}

	i.canalMapa <- acao

	if sucesso := <-resposta; sucesso {
		i.x, i.y = novoX, novoY
	}
}

func distancia(x1, y1, x2, y2 int) int {
	return int(math.Abs(float64(x1-x2)) + math.Abs(float64(y1-y2)))
}
