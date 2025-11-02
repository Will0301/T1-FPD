# Jogo de Terminal em Go

Este projeto é um pequeno jogo desenvolvido em Go que roda no terminal usando a biblioteca [termbox-go](https://github.com/nsf/termbox-go). O jogador controla um personagem que pode se mover por um mapa carregado de um arquivo de texto.

## Como funciona

- O mapa é carregado de um arquivo `.txt` contendo caracteres que representam diferentes elementos do jogo.
- O personagem se move com as teclas **W**, **A**, **S**, **D**.
- Pressione **E** para interagir com o ambiente.
- Pressione **ESC** para sair do jogo.

### Controles

| Tecla | Ação              |
|-------|-------------------|
| W     | Mover para cima   |
| A     | Mover para esquerda |
| S     | Mover para baixo  |
| D     | Mover para direita |
| E     | Interagir         |
| ESC   | Sair do jogo      |

## Como executar

1. Instale o Go (1.21 ou superior) e clone este repositório.
2. As dependências já estão descritas no `go.mod`; basta baixar executando `go mod download` uma vez.
3. Certifique-se de ter os arquivos de mapa (`mapa.txt`, `maze.txt`) no diretório raiz.
4. Inicie o servidor RPC em um terminal separado:

```powershell
go run ./cmd/servidor
```

5. Em outro terminal, execute o jogo localmente. Opcionalmente, passe o nome do arquivo de mapa como primeiro argumento:

```powershell
go run . [nome-do-mapa.txt]
```

Também é possível gerar binários usando `go build`, `make` (Linux) ou `build.bat` (Windows) se preferir.

## Estrutura do projeto

- main.go — Ponto de entrada e loop principal
- interface.go — Entrada, saída e renderização com termbox
- jogo.go — Estruturas e lógica do estado do jogo
- personagem.go — Ações do jogador
- cmd/servidor/main.go — Servidor RPC usado para sincronizar múltiplos clientes
- internal/rpcproto — Tipos compartilhados usados no protocolo RPC


