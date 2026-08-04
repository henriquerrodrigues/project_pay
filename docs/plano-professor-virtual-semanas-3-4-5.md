# 🎓 Professor Ponteiro — Plano de Aprendizagem (Semanas 3, 4 e 5)

> Documento vivo. Atualize a seção "Onde paramos" a cada sessão de estudo — é o que faz a repetição espaçada funcionar de verdade.

## 1. Quem é o Professor Ponteiro 🧑‍🏫👉

Um professor virtual excêntrico, caloroso e cientificamente teimoso: ele só ensina do jeito que a ciência cognitiva mostrou que funciona — e nunca do jeito que "parece" mais rápido, mas não gruda.

| Comportamento do professor | Base científica | Como aparece na prática |
|---|---|---|
| Quebra tudo em pedaços pequenos, um de cada vez | **Cognitive Load Theory** (Sweller, 1988) — memória de trabalho satura com poucos elementos novos simultâneos | Nenhuma explicação chega em bloco único; vem em partes numeradas (Parte 1 de N) |
| Mostra o código pronto explicado ANTES de pedir pra você repetir sozinho | **Worked-example effect** (Sweller & Cooper, 1985) | Primeiro exemplo de cada padrão vem comentado a fundo; o próximo similar você tenta sozinho |
| Faz pergunta antes de avançar pro próximo conceito | **Testing effect / active recall** (Roediger & Karpitke, 2006) | "Antes de seguir, você entende por que X?" |
| Retoma conceito já visto em contexto novo, de tempos em tempos | **Curva do esquecimento + repetição espaçada** (Ebbinghaus, 1885; Cepeda et al., 2006) | Quizzes de revisão no início de cada semana nova, puxando conceito de semana anterior |
| Pede pra você explicar com suas palavras quando há dúvida | **Técnica de Feynman** (popularizada por Richard Feynman) | "Explica pra mim como se eu não soubesse nada de Go" |
| Erro vira explicação, nunca julgamento | **Mindset de crescimento** (Dweck, 2006) | Feedback sempre no formato: o que aconteceu → por que não é o esperado → como pensar diferente |
| Alterna tipo de desafio (código, pergunta de design, quiz relâmpago) | **Interleaving** (Rohrer & Taylor, 2007) — misturar tipos de prática fixa melhor que bloco único do mesmo tipo | Cada "missão" abaixo mistura: 1 implementação + 1 pergunta de design + 1 mini-quiz |

Emojis aparecem para marcar **tipo de conteúdo**, não decoração aleatória — isso é consistente com o requisito de "visual e divertido", mas também funciona como sinalização (facilita escanear a seção certa):

- 🧩 conceito novo · 🛠️ mão na massa · ❓ pergunta de verificação · 🎯 desafio/quiz · ✅ marco batido · ⚠️ armadilha comum · 🏆 celebração

---

## 2. Onde paramos — revisão honesta (26/07/2026)

Isso é o que **já existe no código**, verificado com `go build ./... && go test ./...` (ambos limpos) na branch `pp02`. Nada aqui é otimista — é o estado real.

### Fundação (Task 1-4 da Semana 3)
- ✅ `apperrors` — tipo `Error`, construtores, `WriteHTTP` com trace_id, testado (4 testes)
- ✅ `traceid` — pacote isolado, testado
- ✅ `middleware.UserIDFromContext` — helper adicionado
- ✅ `middleware.Auth` e `handlers/auth.go` migrados pro contrato `apperrors` (token ausente/inválido, credenciais inválidas, erro ao gerar token) — Missão 3.4 fechada. Continua sem uma camada `services/auth` própria (o plano técnico original só previa migrar middleware + handler pro contrato de erro nesta task; `VerifyCredentials`/`CreateToken` seguem no repository)
- ✅ Middleware `RequestLogger` — trace ID (UUID v4) por request, header `X-Trace-Id`, log estruturado via `slog` (method/path/status/duration_ms/trace_id), `slog.SetDefault` com `JSONHandler` ligado no `main.go` — Missão 3.5 fechada, testado (2 testes)

### Service de `users` (Task 5)
- ✅ Completo: `services/users`, testes (6), handler migrado, wiring no `main.go`

### Service de `accounts` (Task 6) — ✅ Missão 3.1 fechada
- ✅ Repository: `Create` recebe `userID`, `GetByID`, `ListByUserID` implementados
- ✅ Model: `CreateAccountRequest` sem `UserID` no body
- ✅ Service completo: `Create` (validação de campos + `accountType` + userID do token) e `GetByID` (ownership 404/403), além do `ListByCurrentUser` já existente
- ✅ `accounts_test.go`: 8 casos (create vincula user do token, accountType inválido, campos obrigatórios, sem user no contexto, getByID dono/outro dono/inexistente/sem user, list só as próprias) — todos passando
- ✅ Handler `accounts.go` migrado pro contrato `apperrors`
- ✅ Wiring + rotas `POST /accounts`, `GET /accounts/{id}`, `GET /accounts` no `main.go`
- 🔲 Faltam só as duas perguntas de verificação da Missão 3.1 (ordem validação token→campos, e NotFound vs Forbidden) — responder antes de considerar a missão 100% internalizada, não só "codada"

### Services de `participants` e `pix_keys` (Task 9) — ✅ Missão 3.2 fechada
- ✅ `services/participants` e `services/pix_keys`: pass-through (sem ownership — catálogo compartilhado), validação de campos obrigatórios, `apperrors.NotFound`/`InvalidInput`/`Internal`
- ✅ `participants_test.go` (2 casos) e `pixkeys_test.go` (2 casos) — todos passando
- ✅ Handlers migrados pro contrato `apperrors`
- ✅ Wiring no `main.go`

### Services de `ledger_entries` e `transfers` (Tasks 7-8) — ✅ Missão 3.3 fechada
- ✅ `services/ledger_entries`: `GetStatement` com ownership da conta (dono → extrato; outro dono → 403; conta inexistente → 404), usa `AccountGetter` (recorte do repository de `accounts`)
- ✅ `ledger_entries_test.go`: 3 casos (dono, outro dono 403, conta inexistente 404) — todos passando
- ✅ `services/transfers`: `Create` (valor > 0, contas diferentes, dono da origem, existência do destino, saldo insuficiente → `apperrors.InsufficientFunds()`) e `GetByID` (visível a quem for dono da origem OU do destino, 403 pra terceiros) — usa as duas interfaces `TransferRepository` + `AccountGetter`
- ✅ `transfers_test.go`: 11 casos — todos passando
- ✅ Handlers migrados pro contrato `apperrors`
- ✅ Wiring no `main.go`

### ✅ Semana 3 fechada
Todos os services em camadas (`users`, `accounts`, `participants`, `pix_keys`, `ledger_entries`, `transfers`), contrato `apperrors` ponta a ponta (incluindo `auth`), e observabilidade básica (`RequestLogger` + `slog`) — `go build ./... && go test ./...` limpos. Falta só o marco de fechamento formal: um `curl` de ponta a ponta (login → criar conta → transferir → ver extrato) confirmando `trace_id` em toda resposta de erro (seção 4 do plano) — isso é smoke manual, não código, e depende do Postgres do compose de pé.

### Ainda não iniciado
- 🔲 Toda a Semana 4 (frontend Next.js)
- 🔲 Toda a Semana 5 (Swagger, `.http`, README, ADR)

**Tradução em progresso:** Semana 3 ~95% (só falta o smoke manual de ponta a ponta), 0% da Semana 4 e 5. Isso é normal nesta altura — o objetivo desta seção é você sempre saber exatamente onde está, sem achismo.

---

## 3. Cronograma até janeiro 🗓️ (recalibrado com o `PLANO_MESTRE.md`)

Fonte de verdade agora é `.github/workflows/plans/PLANO_MESTRE.md` — o plano mestre de 12 meses, premissa de 1h/dia. Semanas 1 e 2 já batidas ✅, Semana 3 em andamento. Projetando o ritmo real observado (cada "semana" = 7 dias corridos):

| Período (estimado) | Semana do plano mestre | Foco | Marco |
|---|---|---|---|
| até ~09/08 | Semana 3 | Auth JWT, `apperrors`, `slog`+trace ID, camadas handler/service/repo | ✅ Backend em camadas, contrato de erro consistente |
| até ~16/08 | Semana 4 | Frontend mínimo (Next.js) | ✅ Login → dashboard → transferência → extrato na tela |
| até ~23/08 | Semana 5 | Swagger, `.http`, README, primeira ADR | ✅ Projeto documentável por outra pessoa sem te perguntar nada |
| até ~30/08 | **Semana 6** | Testes de carga com k6 — **a mais importante da Fase 1** | ✅ `docs/patologias.md` com ≥4 patologias reproduzidas |
| até ~06/09 | Semana 7 | SQL pesado (window functions, CTE, EXPLAIN, índices, keyset pagination) | ✅ ADR de paginação + endpoint de relatório novo |
| até ~13/09 | Semana 8 | CI/CD (GitHub Actions, lint, Docker build, Makefile, cobertura) | ✅ Todo commit sabe em 2min se quebrou algo |
| até ~27/09 | Semanas 9-10 | Observabilidade básica (Prometheus, Grafana, OTel/Jaeger) | ✅ Dashboard + trace ponta a ponta de uma transferência |
| até ~11/10 | Semanas 11-12 | Polimento + `docs/fase-1-relatorio.md` | 🏆 **Fim da Fase 1** — semente do TCC |
| meados out. → janeiro | Fase 2 (idempotência, locking, outbox, RabbitMQ, Redis) — **início** | 📖 detalhado quando chegarmos lá, como o próprio plano mestre promete | 🔲 progresso parcial, sem pressa de terminar |

**Sendo direto com você (o professor não empurra poeira pra baixo do tapete):** a Fase 1 inteira cabe confortavelmente antes de janeiro nesse ritmo — isso é o que importa pro seu pedido original. As Fases 2-4 somam mais 9 meses no próprio plano mestre; não faz sentido fingir que "ficam prontas" até janeiro também. O que fazemos é começar a Fase 2 com a mesma qualidade, sem virar corrida.

---

## 4. Jornada — Semana 3 (o que falta)

### 🧩 Missão 3.1 — Fechar `accounts`
- 🛠️ Implementar `Create` e `GetByID` no `services/accounts` com as 6 regras já discutidas
- 🛠️ Escrever `accounts_test.go` (7 casos: create ok, userID do token, campos obrigatórios, accountType inválido, getByID dono/outro dono/inexistente)
- 🛠️ Migrar `handlers/accounts.go` + wiring + rota `GET /accounts`
- ❓ Antes de fechar: por que `Create` valida token E campos obrigatórios ANTES de chamar o repository, nessa ordem e não invertida?
- 🎯 Quiz relâmpago: qual a diferença entre retornar `apperrors.NotFound` e `apperrors.Forbidden` em termos de "enumeration attack"? (Você já viu essa pergunta — é repetição espaçada de propósito.)

### 🧩 Missão 3.2 — `participants` e `pix_keys` (pass-through, sem ownership)
- 🛠️ Replicar o molde de `users`, mas SEM checagem de dono (participante e chave pix não têm "dono" direto, são catálogo compartilhado)
- ❓ Por que esses dois services são mais simples que `accounts`? O que estrutural neles justifica isso?

### 🧩 Missão 3.3 — `ledger_entries` e `transfers` (dependem de `accounts`)
- 🛠️ `ledger_entries`: extrato só do dono da conta
- 🛠️ `transfers`: valida valor > 0, contas diferentes, dono da conta origem, existência da conta destino, saldo insuficiente → `apperrors.InsufficientFunds()`
- ⚠️ Armadilha clássica: `transfers` precisa de **duas** interfaces de repository (a própria + `AccountGetter`) — não confunda isso com dependência circular de pacote

### 🧩 Missão 3.4 — `auth` (deixado por último, como planejado)
- ❓ Retomando a pergunta em aberto: `VerifyCredentials` devolve `bool` — isso vira erro tipado dentro do repository ou só no service?
- 🛠️ Implementar `services/auth`, migrar `middleware.Auth` e `handlers/auth.go` pro contrato `apperrors`

### 🧩 Missão 3.5 — Observabilidade
- 🛠️ `middleware/request_logger.go` (trace ID + log estruturado), ligar `slog.SetDefault` no `main.go`
- 🎯 Quiz: por que o `statusRecorder` precisa embrulhar o `http.ResponseWriter` em vez de só ler o status depois?

**✅ Marco de fechamento da Semana 3:** `go build ./... && go test ./...` limpo, e um `curl` de ponta a ponta (login → criar conta → transferir → ver extrato) devolvendo trace_id em toda resposta de erro.

---

## 5. Jornada — Semana 4 (frontend mínimo)

Baseada no plano técnico já existente em `docs/superpowers/plans/2026-07-13-semana-4-frontend.md` (10 tasks: seed, scaffold Next.js, helpers de fetch/dinheiro, auth com cookie httpOnly, login/signup, dashboard, dialogs de conta/transferência, extrato, fluxo completo). O Professor Ponteiro vai te guiar por elas na mesma cadência de partes pequenas + pergunta de verificação, quando você chegar lá — não antecipamos código agora porque a Semana 3 ainda não fechou (evita context switching, que atrapalha aprendizado segundo a própria interleaving research: alternar é bom entre **conceitos relacionados**, não entre **projetos inacabados**).

**❓ Pergunta de progressão pra quando chegar aqui:** você prefere aprender a integração com cookie httpOnly primeiro, ou o scaffold do Next.js? Isso muda a ordem sugerida das tasks.

---

## 6. Jornada — Semana 5 (documentação e Swagger)

Baseada em `docs/superpowers/plans/2026-07-13-semana-5-docs.md` (swaggo, `.http`, README, primeira ADR). Fica pra depois da Semana 4 fechar — documentar uma API que ainda muda de forma é retrabalho.

---

## 7. Jornada — Semana 6 (testes de carga — a hora da verdade) 🎯

O próprio plano mestre marca essa como **a semana mais importante da Fase 1**. Objetivo: provar cientificamente que o sistema ingênuo quebra — sem isso, você não tem argumento real pra Fase 2.

### 🧩 Missão 6.1 — k6, primeiro contato
- 🛠️ Escrever o primeiro script k6: 100 RPS de criação de transferências por 1 minuto
- ❓ O que você espera que aconteça, **antes** de rodar? Escreve a hipótese num comentário no script — depois compara com o resultado real (isso é o método científico, não só teste de carga)

### 🧩 Missão 6.2 — Race condition proposital
- 🛠️ Cenário: 50 requests simultâneas de débito na MESMA conta
- 🎯 Quiz preditivo: sem rodar ainda, você acha que o saldo final vai ficar negativo, positivo demais, ou correto? Por quê, em termos de "lost update"?
- ⚠️ Você vai ver saldo negativo de verdade. Isso não é bug seu — é o sistema ingênuo se comportando exatamente como deveria (mal). Comemore o "fracasso", ele é o dado.

### 🧩 Missão 6.3 — Duplicação por retry
- 🛠️ Simular cliente reenviando a mesma transferência 10x (retry ingênuo)
- ❓ Qual conceito da teoria explica isso — "at-least-once delivery" sem idempotência do lado do servidor?

### 🧩 Missão 6.4 — Documentar as patologias
- 🛠️ `docs/patologias.md`: para cada patologia — o que aconteceu, por que (mecanismo técnico), qual teoria explica (race condition / lost update / falta de idempotência)
- ❓ Técnica de Feynman: explica a `docs/patologias.md` em voz alta pra alguém (ou pra você mesmo gravando áudio) como se a pessoa não soubesse nada de banco de dados

### 🧩 Missão 6.5 — Isolation levels do Postgres
- 🛠️ Reproduzir o saldo negativo, depois testar `SERIALIZABLE` e ver o problema sumir (e a performance cair — meça os dois)
- 🎯 Quiz: por que `SERIALIZABLE` resolve o saldo negativo, mas você não vai usar ele como solução definitiva na Fase 2? (dica: o plano mestre já entrega a resposta na Fase 2 — locking explícito)

**✅ Marco:** `docs/patologias.md` com no mínimo 4 patologias, cada uma com reprodução + métrica + teoria. Isso é celebração de verdade — poucos alunos quebram o próprio sistema de propósito e documentam cientificamente. 🏆

---

## 8. Jornada — Semana 7 (SQL pesado) 🧩

Você já sinalizou SQL como ponto fraco — essa semana é desenhada pra doer um pouco e sair diferente do outro lado.

- 🛠️ Window functions (`OVER`, `PARTITION BY`) pra saldo histórico no extrato
- 🛠️ CTEs e CTEs recursivos num relatório
- 🛠️ `EXPLAIN ANALYZE` nas queries atuais — achar full table scans
- 🛠️ Índices certos + re-medir
- 🛠️ Endpoint novo: `GET /reports/daily-volume` com **keyset pagination** (não offset)
- 🛠️ ADR: keyset vs offset, com números

❓ **Antes de cada `EXPLAIN ANALYZE`:** escreve sua previsão de qual linha do plano vai ser o gargalo. Comparar previsão vs realidade é o "active recall" aplicado a leitura de plano de execução, não só a código.

⚠️ Armadilha clássica de keyset pagination: gente tenta usar como cursor um campo não-único (ex: só `created_at`, sem `id` de desempate) — dois registros com timestamp igual quebram a paginação. Guarda isso antes de implementar.

**✅ Marco:** o próprio plano mestre diz — "você é uma pessoa diferente em SQL ao final desta semana". Vamos testar isso com um quiz de fechamento quando chegar lá.

---

## 9. Jornada — Semana 8 (CI/CD básico) 🛠️

- 🛠️ GitHub Actions: lint (`golangci-lint`) + testes em cada PR
- 🛠️ Build de imagem Docker no pipeline + push pro GHCR
- 🛠️ Pre-commit hooks locais (`gofmt`, `golangci-lint`, testes rápidos)
- 🛠️ `Makefile` (`make run`, `make test`, `make migrate`, `make load-test`)
- 🛠️ Cobertura de testes ~70% em `internal/` (guia, não dogma)

❓ Por que rodar lint e teste **no PR** (antes do merge) é diferente, em termos de custo de correção, de rodar só depois do merge na branch principal? (Relação direta com o "quanto antes o erro aparece, mais barato corrigir" — princípio conhecido de qualidade de software.)

**✅ Marco:** você commita, e em 2 minutos sabe se quebrou algo.

---

## 10. Jornada — Semanas 9-10 (Observabilidade básica) 🧩

- 🛠️ Métricas Prometheus: contador de requests, histograma de latência, gauge de transferências em andamento
- 🛠️ Prometheus + Grafana no Compose, dashboard básico
- 🛠️ Traces com OpenTelemetry + Jaeger — ver uma transferência ponta a ponta
- 🛠️ ADR da escolha de stack de observabilidade
- ✅ Folga estruturada nos dias finais (dia 67-70 do plano mestre) — revisão geral, refatorações pontuais, descanso. Isso está no plano mestre por um motivo científico: consolidar sem input novo é parte do aprendizado, não desperdício de tempo.

❓ Diferença prática entre **métrica** (Prometheus), **trace** (Jaeger/OTel) e **log** (slog + trace_id que você já tem desde a Semana 3)? Se você não souber responder de cabeça ainda, ótimo — essa é a pergunta que abre a Semana 9, não fecha.

---

## 11. Jornada — Semanas 11-12 (Polimento + Relatório de Fase 1) 🏆

- 🛠️ Casos de borda, mensagens de erro melhores, README impecável
- 🛠️ `docs/fase-1-relatorio.md`: o que foi feito, decisões, patologias encontradas, métricas atuais, plano da Fase 2 — **primeira semente do TCC**

**✅ Marco final da Fase 1:** sistema funcionando, quebrado de propósito, documentado. Pronto pra Fase 2. Esse é o ponto de "certificação informal" — vale parar e reconhecer o tamanho do que foi construído antes de seguir.

---

## 12. Visão geral — Fases 2, 3 e 4 (detalhamento fica pra quando chegarmos lá)

O próprio `PLANO_MESTRE.md` só tem granularidade semana-a-semana pra Fase 1 — Fases 2-4 são blocos temáticos, e o Professor Ponteiro vai quebrar em missões só quando você chegar perto (evita carga cognitiva de planejar 9 meses à frente com informação que ainda vai mudar). Resumo do que vem:

| Fase | Meses | Tema central |
|---|---|---|
| **2 — O Motor Correto** | 4-6 | Idempotência, locking (pessimista vs otimista), outbox pattern, separação em `pagar-api` + `pagar-processor` via RabbitMQ, Redis (cache + lock distribuído), re-teste de carga comparando com Fase 1 |
| **3 — O Ecossistema** | 7-9 | Deploy em Hetzner, múltiplos "bancos" simulados, SPI simulado (roteamento), padrão Saga, observabilidade distribuída, chaos testing |
| **4 — Inteligência + TCC** | 10-12 | Detecção de fraude (isolation forest ou similar) em serviço Python consumindo RabbitMQ, integração com bloqueio/confirmação, **escrita do TCC** |

🎯 Quiz de longo prazo (revisita essa pergunta a cada início de fase): por que o plano mestre insiste em "profundidade sobre amplitude" — poucos endpoints que aguentam porrada, em vez de muitas features rasas? Sua resposta hoje pode (e deve) mudar depois da Semana 6, quando você sentir na pele o que "aguentar porrada" custa pra construir.

---

## 13. Armadilhas já mapeadas pelo plano mestre ⚠️

O plano mestre já antecipa os riscos mais prováveis — o Professor Ponteiro só traduz em comportamento prático:

| Risco | O que fazer quando acontecer |
|---|---|
| Travar numa semana sem avançar | No fim do dia 7 da semana, avança mesmo com débito técnico — o débito vai pro `BACKLOG.md`, não fica te bloqueando |
| "Síndrome do tutorial" (copiar sem entender) | É por isso que cada missão aqui tem pelo menos uma ❓ de "explica com suas palavras" antes de seguir |
| Perfeccionismo | Commit feio no fim da semana > commit perfeito que nunca sai. Revisa depois, friamente |
| Vida acontece (faculdade, estágio) | 1 dia ruim é ok; 3 dias ruins é alarme — mas na volta você **não** tenta recuperar o tempo perdido, só retoma de onde parou |
| Comparação com projetos rasos de outras pessoas | Profundidade > amplitude é o próprio critério de sucesso deste projeto, não velocidade aparente |

---

## 14. Biblioteca do professor 📚 (leituras oficiais do plano mestre)

Curadoria enxuta, uma prioridade por fase — não é pra ler tudo de uma vez:

- **Fase 1 (agora):** *Designing Data-Intensive Applications* (Kleppmann), caps. 1-3 e 7 — **prioridade máxima**, ~10 páginas/semana; *Use The Index, Luke* (leitura completa na Semana 7); Effective Go + doc oficial do Go
- **Fase 2:** DDIA caps. 5, 8, 9, 11; Fowler sobre Unit of Work / Identity Map / Outbox; docs do Postgres sobre MVCC e isolation levels
- **Fase 3:** DDIA caps. 8, 9, 11 (releitura — vai fazer mais sentido); *Site Reliability Engineering* (Google, grátis online); docs OpenTelemetry
- **Fase 4:** *Hands-On Machine Learning* (Géron), capítulos de detecção de anomalia; papers específicos do tema do TCC

❓ Repetição espaçada aplicada à leitura: quando você reler DDIA cap. 8-9 na Fase 3, antes de começar, tenta lembrar (sem abrir o livro) o que anotou da primeira leitura na Fase 2. Isso sozinho já é um exercício de recall válido.

---

## 15. Como retomar uma sessão com o Professor Ponteiro

No início de cada sessão nova de estudo, comece pedindo: **"revisa onde paramos"** — isso força a atualização da Seção 2 antes de qualquer código novo, e é o gatilho de repetição espaçada que faz o conteúdo grudar. Se uma sessão terminar com algo funcionando, atualize os ✅/🛠️/🔲 desta seção antes de fechar — é rápido e paga dividendo enorme na sessão seguinte.

---

## 16. Resumo geral — tecnologias, problemas e material de TCC 📦

### 🛠️ Stack por fase (o que entra quando, e por quê)

| Tecnologia | Entra na | Problema que resolve |
|---|---|---|
| **Go** | Fase 1 | Linguagem principal do motor — performance previsível, concorrência nativa |
| **PostgreSQL 16+** (`pgx`, não `database/sql` puro) | Fase 1 | Persistência transacional; `pgx` porque é mais eficiente e idiomático pra Postgres especificamente |
| **`net/http` (Go 1.22+) + camadas handler/service/repository** | Fase 1 (Semana 1 e 3) | Organização do código — separa "fala HTTP", "regra de negócio" e "fala com banco", cada um testável isoladamente |
| **`apperrors` (pacote próprio) + `slog` + trace ID** | Semana 3 | Contrato de erro consistente que nunca vaza detalhe interno; correlação de log por requisição |
| **JWT** | Semana 3 | Autenticação; cada conta passa a pertencer a um usuário (ownership) |
| **Next.js + Tailwind + shadcn/ui** | Fase 1 (mínimo, Semana 4) → Fase 2 (decente) | Interface pra "ver o sistema funcionando", sem sofrer com CSS cru |
| **k6** | Semana 6 (mas testado desde a Fase 1!) | Provar sob carga real que concorrência sem proteção gera saldo negativo e duplicação |
| **golangci-lint + GitHub Actions + Docker + Makefile** | Semana 8 | Automação — saber em minutos se um commit quebrou algo |
| **Prometheus + Grafana + OpenTelemetry + Jaeger** | Semanas 9-10 (básico) → Fase 3 (completo) | "Ver" o sistema: métricas, dashboards, traces ponta a ponta |
| **Redis** | Fase 2 | Cache de leitura de saldo + lock distribuído entre serviços |
| **RabbitMQ** | Fase 2 | Mensageria — desacopla `pagar-api` de `pagar-processor`, viabiliza outbox pattern |
| **Docker + Docker Compose** | Fase 1 (banco local) → Fase 3 (deploy) | Ambiente reprodutível local, depois produção |
| **Hetzner (VPS)** | Fase 3 | Hospedagem real, primeira vez que o sistema sai da sua máquina |
| **Python + scikit-learn (isolation forest ou similar)** | Fase 4 | Detecção de fraude/anomalia consumindo eventos do RabbitMQ |

### 🎯 Problemas reais que o projeto resolve (não é feature por feature — é patologia por patologia)

1. **Corrupção de saldo sob concorrência** (lost update / race condition) — provado na Semana 6, corrigido na Fase 2 com locking (pessimista `SELECT FOR UPDATE` vs otimista por versão)
2. **Duplicação de transferência em retry** (ausência de idempotência) — provado na Semana 6, corrigido na Fase 2 com chave de idempotência + tabela de deduplicação
3. **Entrega de eventos não confiável entre serviços** — resolvido com outbox pattern (Fase 2) e, depois, roteamento entre "bancos" simulados com garantia de entrega (Fase 3)
4. **Consistência distribuída em falhas parciais** (um serviço cai no meio de uma operação multi-etapa) — endereçado com padrão Saga e compensação (Fase 3)
5. **Opacidade operacional** ("por que isso está lento/quebrado e eu não sei onde") — resolvido com observabilidade (métricas, logs correlacionados por trace ID, traces distribuídos) começando na Semana 3 e amadurecendo até a Fase 3
6. **Detecção de comportamento anômalo em tempo real** (fraude) — Fase 4, via modelo de ML consumindo o stream de eventos
7. **Paginação que degrada com o crescimento da tabela** (offset pagination) — resolvido com keyset pagination na Semana 7

### 📖 Material que já nasce virando TCC

- **Toda ADR** (Architecture Decision Record) escrita ao longo do projeto — pgx vs `database/sql`, keyset vs offset, locking pessimista vs otimista, stack de observabilidade, e mais — vira seção de "decisões e trade-offs" do TCC
- **`docs/patologias.md`** (Semana 6) — a prova empírica do problema que o TCC resolve; sem isso não tem motivação científica pro trabalho
- **`docs/fase-1-relatorio.md`, e os relatórios equivalentes de Fase 2 e 3** — comparativos "antes/depois" com métricas reais (throughput, latência, taxa de erro sob carga) — é a espinha dorsal dos capítulos de resultados
- **Os dois temas de TCC já sugeridos no plano mestre**, pra decidir com calma até a Fase 4:
  1. *"Estratégias de garantia de idempotência e consistência em sistemas de pagamentos instantâneos: uma análise comparativa"* — mais alinhado ao trabalho de Fases 1-3
  2. *"Detecção de anomalias em redes de pagamentos instantâneos via [técnica escolhida]"* — puxa mais peso pro componente de ML da Fase 4
- **O chaos testing da Fase 3** — vira o capítulo de "resiliência e recuperação de falhas", com dados reais de como o sistema se comportou sob falha induzida

❓ Pergunta de fechamento pra você guardar: à medida que for escrevendo ADRs, vale já ir marcando (num campo extra, tipo `TCC: sim/não/talvez`) quais decisões têm peso acadêmico suficiente pra virar seção do trabalho — evita, lá na Fase 4, ter que reler 30+ ADRs pra descobrir quais interessam.
