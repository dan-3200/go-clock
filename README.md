# Go Clock

Pomodoro no terminal com Bubble Tea, Bubbles e Lip Gloss. Foco, pausas, configurações e métricas locais, sem conta ou serviço remoto.

Com Go 1.27.2 ou superior:

```sh
go run .
go build -o go-clock .
./go-clock
```

Use um terminal interativo. O app abre em tela alternativa e restaura o terminal ao sair.

| Atalho | Ação |
| --- | --- |
| Espaço / Enter | Iniciar, pausar ou retomar |
| r | Reiniciar o período atual, deixando-o parado |
| n | Avançar sem contabilizar o foco interrompido |
| s | Abrir Settings |
| q / Esc / Ctrl+C | Sair |
| Tab / Shift+Tab / ↑ / ↓ | Selecionar campo em Settings |
| Enter / Esc | Salvar / cancelar Settings |

Os padrões são 25 minutos de Pomodoro, 5 de Short Break, 15 de Long Break e uma pausa longa a cada 4 focos concluídos. Todos são configuráveis. Durações aceitam inteiros de 1 a 10080 minutos; o intervalo da pausa longa aceita 1 a 1000 focos. Mudanças valem para o próximo período; reiniciar preserva a duração do período atual. O relógio continua ativo com Settings aberto.

Cada conclusão seleciona o próximo período, que aguarda início manual. Somente focos concluídos incrementam ciclos e horas: pausas, sessões interrompidas e tempo pausado não contam. Ao sair de uma pausa longa, concluída ou avançada manualmente, começa um novo ciclo. Ao reabrir, o app inicia um Pomodoro parado e preserva as métricas e configurações; o cronômetro em andamento não é restaurado.

Os dados ficam em `go-clock/data.json` dentro de `os.UserConfigDir()`:

- Linux: `$XDG_CONFIG_HOME/go-clock/data.json`, ou `~/.config/go-clock/data.json`.
- macOS: `~/Library/Application Support/go-clock/data.json`.
- Windows: `%AppData%\go-clock\data.json`.

A gravação usa arquivo temporário e substituição após sincronização. Erros de gravação aparecem na interface e são tentados novamente ao sair. Um arquivo inválido impede a abertura, preservando seus dados para recuperação manual. As métricas não possuem botão de reset.

Alertas usam `beeep` para notificação desktop. Linux exige sessão gráfica com serviço de notificações/D-Bus. O som usa `canberra-gtk-play -i complete` no Linux (instale o utilitário e um tema de sons), `afplay` com Glass no macOS e PowerShell/SystemSounds no Windows. Se o som nativo falhar, o app envia a campainha do terminal, que depende das preferências do terminal. Falhas de integração exibem aviso, sem interromper o Pomodoro. Notificação e som executam fora do loop da interface. SSH e terminais sem desktop mantêm o aviso visual.

A arquitetura segue `main.go` → `internal/run.go`, com `engine` para eventos e cronômetro, `models` para dados, `components` para telas, `storage` para JSON e `notifications` para integrações do sistema.

```sh
go test ./...
go vet ./...
```
