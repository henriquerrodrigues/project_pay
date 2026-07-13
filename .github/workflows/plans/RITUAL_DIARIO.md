# Ritual Diário — 1 hora bem investida

> O inimigo não é falta de tempo. O inimigo é a hora que vira "ler tutorial e não codar nada".

## A regra de ouro

**Toda sessão de 1h termina com código commitado ou documento escrito.** Sempre. Sem exceção.

Se você passou 1h só "estudando", você não progrediu — você se entreteve com programação. Estudo puro pode existir, mas **não dentro da sua hora do projeto**. Estudo vai pra outro momento (almoço, deslocamento, fim de semana). A hora do projeto é hora de *produzir*.

## Estrutura padrão de 60 minutos

```
[ 5 min ] Aquecimento  — abre PLANO_MESTRE.md, lê a tarefa do dia, abre o repo.
[ 50 min ] Foco        — Pomodoro de 25+5+25 ou 50 min direto. Sem celular.
[  5 min ] Fechamento  — commit, anota no DIARIO.md o que fez e onde parou.
```

O fechamento é sagrado. O `DIARIO.md` tem três linhas por dia:

```markdown
## 2026-05-01 — Dia 4 — Docker Compose com Postgres
Fiz: subiu compose, conectei via pgx, primeiro SELECT funcionando.
Próximo: criar primeira migration com golang-migrate.
Travas: nenhuma.
```

Esse diário, no fim do projeto, é ouro. Pra você mesmo, pra mostrar pra recrutador, pra usar no TCC.

## Os três modos de sessão

Cada dia você está em um dos três modos. Reconheça qual é antes de começar.

### Modo BUILD (constrói feature)
A maioria dos dias. Sabe o que tem que fazer, faz. Mede progresso em "feature funcionando + teste passando".

### Modo LEARN (aprende pra construir)
Quando bateu numa coisa que você não sabe (ex.: "como funciona SELECT FOR UPDATE"). Regra: **leitura tem que terminar em código.** Leu sobre lock pessimista? Os últimos 20 minutos da sessão são você aplicando isso num teste do projeto. Mesmo que mal feito. A teoria sem código não fixa.

### Modo REFACTOR/DOC (consolida)
Sextas e finais de semana. Limpa código, escreve ADR, atualiza README, revisa o que fez.

## Anti-padrões pra reconhecer e cortar

| Sintoma | O que está acontecendo | Antídoto |
|---|---|---|
| Vê 4 vídeos de YouTube sobre Go | Procrastinação produtiva | Fecha o YouTube, abre o editor |
| Refaz a mesma feature pela 3ª vez | Perfeccionismo | Commita feio, segue. Refatora depois com motivo. |
| Passou a sessão lendo Hacker News sobre fintech | Pesquisa que vira distração | Bloqueia HN no horário da sessão |
| "Vou só configurar o ambiente direito antes de começar" | Yak shaving | Define um stop loss: 2 dias no setup, depois é o que tem |
| Reescreve em outra linguagem/framework | Síndrome do framework novo | Você escolheu Go. Acabou. Não muda no meio. |

## Regras semanais

- **Sexta-feira é sagrada pra fechamento de semana.** 30 min do código, 30 min revisando o `DIARIO.md` da semana e atualizando o `PLANO_MESTRE.md` se a realidade desviou.
- **Domingo (ou um dia da semana fixo) é o dia de não-tocar.** Cérebro precisa de off. Sem culpa.
- **Toda semana acaba com commit que documenta o estado.** Mesmo que o código esteja meio quebrado. Pode ser num branch.

## Regras mensais

- **No último dia do mês, escreva uma "retrospectiva" em uma página.** O que aprendeu, o que travou, o que mudou no plano. Cinco minutos disso por mês = documento de uma página no fim do ano = capítulo do TCC e narrativa pra recrutador.
- **Recompensa por mês completo.** Sério. Você tá investindo 30h/mês — comemora. Algo que você gosta, sem culpa.

## Métrica única de sucesso da semana

No domingo (ou dia de off), responde uma pergunta:

> **"Eu sei algo hoje que não sabia segunda?"**

Se sim, semana boa, mesmo que pouco código tenha saído.
Se não, alarme: a próxima semana muda algo na rotina.

## Sobre estudar fora da hora-projeto

Sua hora é pra construir. Mas DDIA, doc de Postgres, etc., são leitura. Onde encaixar?

- **Manhã, antes do trabalho:** 20-30 min de leitura. Café e livro.
- **Almoço:** 20 min de doc.
- **Ônibus/transporte:** podcasts ou audiobook.
- **Sábado:** 1-2h de leitura aprofundada.

Não é obrigatório. Mas faz uma diferença grande sem comer da hora-projeto.

## Quando consultar a IA (e quando não)

Você tem ferramentas de IA boas. Use bem.

**Use:**
- Para revisar uma decisão arquitetural ("estou pensando em X por causa de Y, faz sentido?")
- Para acelerar leitura ("me explica MVCC do Postgres em 5 minutos")
- Para revisar código ("achei isso aqui esquisito, o que tá errado?")
- Para gerar testes de carga, scripts, código boilerplate

**Não use:**
- Para escrever a feature inteira por você. Isso anula o aprendizado. A regra: a IA explica conceitos, você escreve o código.
- Para pular o passo de sentir a dor. Se você ainda não rodou o teste de carga e viu o saldo negativo, não pergunta pra IA "como evitar saldo negativo". Sente a dor primeiro, depois pergunta.

A diferença entre quem aprende profundamente com IA e quem fica raso é essa: a profunda usa IA pra explicar e revisar; a rasa usa IA pra escrever no lugar dela.
