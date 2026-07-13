# Project Pay — Plano Mestre de 12 Meses

> Motor de pagamentos instantâneos inspirado no PIX. Foco em throughput, idempotência e consistência distribuída.

## Princípios que guiam o plano

1. **A dor é o professor.** Você sempre constrói a versão ingênua primeiro, prova que ela quebra, depois refatora. Cada refatoração vira aprendizado profundo (e capítulo de TCC).
2. **Sempre tem algo rodando.** Ao final de cada fase, você tem um sistema funcional, mais sofisticado que o anterior. Nunca um sistema pela metade.
3. **Profundidade sobre amplitude.** Poucas funcionalidades, cada uma tocando problemas reais. Não é uma fintech de mentira com 50 features — é um motor de poucos endpoints que aguenta porrada.
4. **Documentar enquanto aprende.** Cada decisão arquitetural vira uma ADR (Architecture Decision Record). No fim do projeto você tem 30+ ADRs que viram a base do TCC.

---

## Stack final (referência)

| Camada | Tecnologia | Quando entra |
|---|---|---|
| Linguagem principal | Go | Fase 1 |
| Banco principal | PostgreSQL 16+ | Fase 1 |
| Cache / locks distribuídos | Redis | Fase 2 |
| Mensageria | RabbitMQ | Fase 2 |
| Frontend | Next.js + Tailwind + shadcn/ui | Fase 1 (mínimo), Fase 2 (decente) |
| Observabilidade | OpenTelemetry + Prometheus + Grafana + Loki | Fase 3 |
| Container | Docker + Docker Compose | Fase 1 |
| CI | GitHub Actions | Fase 1 |
| Hospedagem | Local até Fase 2; Hetzner CX22 a partir da Fase 3 | Fase 3 |
| ML (fraude) | Python + scikit-learn (Fase 4) | Fase 4 |
| Testes de carga | k6 | Fase 1 (já!) |

---

## Visão das 4 fases

### Fase 1 (mês 1-3) — O Motor Ingênuo
**Objetivo:** ter um sistema de pagamentos funcionando do jeito errado, com testes que provam que está errado.
**Entregável:** API REST com contas, transferências, extrato. Frontend mínimo. Suite de testes de carga que demonstra falhas de concorrência e duplicação. Documento de "patologias encontradas".

### Fase 2 (mês 4-6) — O Motor Correto
**Objetivo:** consertar tudo que doeu na Fase 1, com argumentos técnicos pra cada escolha.
**Entregável:** mesma API, agora com idempotência, locking adequado, outbox pattern, separação em 2 serviços (API + processador) comunicando via RabbitMQ. Mesmos testes da Fase 1 passando. Comparativo "antes/depois" com métricas.

### Fase 3 (mês 7-9) — O Ecossistema
**Objetivo:** simular o ecossistema PIX — múltiplos bancos, rede de liquidação, falhas.
**Entregável:** N instâncias do serviço (cada uma um "banco"), uma "rede" central que roteia mensagens entre eles, observabilidade completa (Grafana com dashboards, traces distribuídos), suite de chaos testing. Deploy em Hetzner.

### Fase 4 (mês 10-12) — Inteligência + TCC
**Objetivo:** camada de detecção de fraude + polimento + escrita do TCC.
**Entregável:** serviço Python consumindo eventos do RabbitMQ, detectando anomalias com isolation forest (ou similar). Sistema todo deployado, documentado, com README impecável. TCC escrito.

---

# Fase 1 detalhada — semana a semana

> Premissa: 1h/dia, 7 dias/semana = 7h/semana. Cada semana tem entregável concreto.

## Semana 1 — Fundação e domínio

**Objetivo:** entender o problema, montar o esqueleto.

- [X] Dia 1: Ler sobre o PIX real — como funciona o SPI, o DICT, mensagens ISO 20022 (visão geral, não decora). Anotar em `docs/dominio.md`.
- [X] Dia 2: Modelar o domínio em papel. Quais entidades? (Conta, Participante, Transferência, Lançamento). Quais estados uma transferência pode ter? Desenhar máquina de estados.
- [X] Dia 3: Setup do repositório no GitHub. Estrutura de pastas Go (recomendo `cmd/`, `internal/`, `pkg/`, evitando `pkg/` se não for biblioteca). README inicial com a visão do projeto.
- [X] Dia 4: Docker Compose com Postgres rodando. Conexão básica do Go ao Postgres usando `pgx` (não `database/sql` puro — `pgx` é melhor pra Postgres).
- [X] Dia 5: Primeira migration de banco — tabelas `accounts` e `transfers`. Use `golang-migrate` ou `goose` pra versionar migrations desde o dia 1.
- [X] Dia 6: Endpoint POST /accounts e GET /accounts/{id}. HTTP usando `chi` ou `net/http` puro do Go 1.22+. Sem autenticação ainda.
- [X] Dia 7: Revisão da semana — escreva no `docs/diario.md` o que aprendeu. Descanse o cérebro também.

**Marco da semana:** repositório criado, banco rodando, dois endpoints funcionando, conseguir criar uma conta via curl.

## Semana 2 — Transferências (versão ingênua)

**Objetivo:** implementar transferência entre contas do jeito mais simples possível, sem se preocupar com nada.

- [X] Dia 8-9: Endpoint POST /transfers. Lê conta origem, debita; lê conta destino, credita. Sem lock, sem idempotência. Tudo numa transação SQL simples.
- [x] Dia 10: GET /accounts/{id}/statement com paginação básica. Aqui você já vai brigar com SQL um pouco — JOIN entre transfers e accounts, ordenação, LIMIT/OFFSET.
- [x] Dia 11-12: Testes unitários e de integração. Use `testify`. Banco de teste via testcontainers ou um Postgres separado no Compose.
- [X] Dia 13: Validações de negócio: saldo insuficiente, conta inexistente, valor inválido, mesma conta origem/destino. Retorne erros HTTP apropriados.
- [X] Dia 14: Revisão. Commit "Fim da Semana 2: transferências funcionando do jeito ingênuo".

**Marco:** você consegue criar duas contas, transferir entre elas, ver o extrato. Tudo errado, mas funciona pro caminho feliz.

## Semana 3 — Autenticação e estrutura de erros

**Objetivo:** colocar a casa em ordem antes de escalar.

- [ ] Dia 15-16: Autenticação via JWT (você já conhece — não vai ser difícil). Endpoint /auth/login. Middleware de autenticação. Cada conta pertence a um usuário.
- [ ] Dia 17: Estrutura de erros consistente. Crie um pacote `errors` no projeto que mapeia erros de domínio para códigos HTTP e mensagens. Padrão: erro nunca vaza detalhe interno pro cliente.
- [ ] Dia 18: Logs estruturados com `slog` (Go 1.21+). Cada request tem um trace ID propagado.
- [ ] Dia 19-20: Refatoração leve — separe handlers (HTTP) de services (regra de negócio) de repositories (acesso a dados). Camadas claras. Isso vai te servir até o fim do projeto.
- [ ] Dia 21: Revisão.

**Marco:** API com autenticação funcional, estrutura de código limpa em camadas.

## Semana 4 — Frontend mínimo

**Objetivo:** ter uma tela onde você vê as coisas acontecendo. Não precisa ser bonito ainda.

- [ ] Dia 22: Setup Next.js (App Router). Configurar conexão com a API.
- [ ] Dia 23-24: Tela de login e cadastro.
- [ ] Dia 25-26: Dashboard mostrando contas e saldo. Botão de transferência.
- [ ] Dia 27-28: Tela de extrato. Tailwind + shadcn/ui pra não sofrer com CSS.

**Marco:** você abre o navegador, faz login, transfere dinheiro entre contas, vê o extrato. Apresentável (mesmo que feio).

## Semana 5 — Documentação e Swagger

**Objetivo:** API documentada de verdade. Você já sabe Swagger — agora aplica bem.

- [ ] Dia 29-30: Swagger gerado a partir de anotações no código com `swaggo`. Servir UI em /swagger.
- [ ] Dia 31: Postman collection ou arquivos .http versionados no repo.
- [ ] Dia 32: README do projeto: visão, como rodar, arquitetura atual (com diagrama), próximos passos.
- [ ] Dia 33-34: Sua primeira ADR (Architecture Decision Record). Tema sugerido: "Por que pgx em vez de database/sql + driver". Crie a pasta `docs/adr/`.
- [ ] Dia 35: Revisão.

**Marco:** projeto documentado o suficiente pra outra pessoa rodar sem te perguntar nada.

## Semana 6 — Testes de carga (a hora da verdade)

**Objetivo:** ESTA É A SEMANA MAIS IMPORTANTE DA FASE 1. Você vai provar que seu sistema quebra.

- [ ] Dia 36: Estudar k6. Escrever o primeiro teste — 100 RPS de criação de transferências, durante 1 minuto, ver o que acontece.
- [ ] Dia 37: Cenário de concorrência: a MESMA conta com 50 requests simultâneas de débito. **Você vai ver saldo negativo.** Anote.
- [ ] Dia 38: Cenário de duplicação: o cliente envia 10x a mesma transferência (simulando retry). **Você vai ver o dinheiro debitado 10x.** Anote.
- [ ] Dia 39-40: Documente cada patologia em `docs/patologias.md`. Para cada uma: o que aconteceu, por que aconteceu, qual conceito teórico explica (ex.: race condition, lost update, ausência de idempotência).
- [ ] Dia 41: Estudar isolation levels do Postgres (READ COMMITTED, REPEATABLE READ, SERIALIZABLE). Reproduzir o problema de saldo negativo, depois testar mudando o nível de isolamento e ver o que acontece (e quanto cai a performance).
- [ ] Dia 42: Revisão. Esta é a hora de comemorar — você QUEBROU seu próprio sistema cientificamente. Isso é o pré-requisito pra construir bem.

**Marco:** documento `docs/patologias.md` com pelo menos 4 patologias documentadas com reprodução, métricas, e teoria.

## Semana 7 — SQL pesado

**Objetivo:** atacar sua maior dificuldade declarada (SQL).

- [ ] Dia 43: Estudo focado — janelas analíticas (`OVER`, `PARTITION BY`). Aplicar pra calcular saldo histórico (saldo após cada transação no extrato).
- [ ] Dia 44: CTEs (`WITH`) e CTEs recursivos. Aplicar em algum relatório.
- [ ] Dia 45: EXPLAIN ANALYZE. Analisar suas queries atuais. Identificar full table scans.
- [ ] Dia 46: Índices. Criar os índices certos. Re-medir EXPLAIN.
- [ ] Dia 47-48: Implementar um endpoint novo desafiador: GET /reports/daily-volume — volume de transferências agregado por dia, com filtros, ordenação, paginação keyset (não offset). Se ficar perdido, pesquisa "keyset pagination".
- [ ] Dia 49: ADR sobre estratégia de paginação (keyset vs offset, com argumentos de performance).

**Marco:** você é uma pessoa diferente em SQL ao final desta semana. Sério.

## Semana 8 — CI/CD básico

**Objetivo:** automação que vai te servir pelos próximos 10 meses.

- [ ] Dia 50: GitHub Actions — pipeline de lint (golangci-lint) + testes em cada PR.
- [ ] Dia 51: Build de imagem Docker no pipeline, push pra GitHub Container Registry.
- [ ] Dia 52-53: Pre-commit hooks locais (gofmt, golangci-lint, go test rápido).
- [ ] Dia 54: Makefile com os comandos comuns (`make run`, `make test`, `make migrate`, `make load-test`).
- [ ] Dia 55-56: Cobertura de testes. Aim for ~70% no `internal/`. Não é dogma, é guia.

**Marco:** você commita, e em 2 minutos sabe se quebrou algo.

## Semanas 9-10 — Observabilidade básica

**Objetivo:** começar a "ver" o sistema.

- [ ] Dia 57-58: Adicionar métricas Prometheus básicas — contadores de requests, histograma de latência, gauge de transferências em andamento.
- [ ] Dia 59-60: Subir Prometheus + Grafana no Docker Compose. Dashboard básico.
- [ ] Dia 61-63: Adicionar traces básicos com OpenTelemetry. Jaeger no Compose. Ver uma transferência ponta a ponta.
- [ ] Dia 64-66: Documentar o que você aprendeu. ADR sobre escolha de stack de observabilidade.
- [ ] Dia 67-70: Folga estruturada — revisão geral do código, refatorações pontuais, descanso.

## Semanas 11-12 — Polimento e relatório de Fase 1

- [ ] Dia 71-77: Polimento. Casos de borda, melhoria de mensagens de erro, README impecável.
- [ ] Dia 78-84: Escrever **Relatório de Fase 1** em `docs/fase-1-relatorio.md`. Conteúdo: o que foi feito, decisões, patologias encontradas, métricas atuais, plano da Fase 2. Esse documento é a primeira semente do TCC.

**Marco final da Fase 1:** sistema funcionando, quebrado de propósito, documentado. Pronto pra Fase 2.

---

# Fases 2, 3 e 4 — visão geral (vou detalhar quando chegarmos lá)

## Fase 2 (mês 4-6) — O Motor Correto

Temas semana a semana (sem detalhe diário ainda — vamos calibrar com a realidade da Fase 1):

- **Semanas 13-14:** Idempotência. Chaves de idempotência, tabela de deduplicação, comportamento sob retry.
- **Semanas 15-16:** Locking. Pessimista (`SELECT FOR UPDATE`) vs otimista (versão na linha). Implementar ambos, comparar, escolher com argumento. ADR.
- **Semanas 17-18:** Outbox pattern. Emitir eventos de domínio garantindo entrega exatamente uma vez. Tabela `outbox`, processo que lê e publica.
- **Semanas 19-20:** Introduzir RabbitMQ. Separar serviço em `pagar-api` (recebe HTTP) e `pagar-processor` (consome fila e processa transferências). Comunicação via fila.
- **Semanas 21-22:** Redis entra. Cache de leitura de saldo (com invalidação consistente — cuidado!). Lock distribuído pra operações entre serviços.
- **Semanas 23-24:** Re-rodar TODOS os testes de carga da Semana 6. Comparar números. Polimento. Relatório de Fase 2.

## Fase 3 (mês 7-9) — O Ecossistema

- **Semanas 25-26:** Provisionar VPS Hetzner. Setup de servidor (firewall, ssh-only-key, Docker, Caddy ou Nginx, certificados Let's Encrypt). Deploy do sistema atual.
- **Semanas 27-28:** Refatorar pra suportar múltiplos "bancos" simulados — cada um uma instância separada do serviço, com seu próprio banco de dados. Conceito de Participant Code.
- **Semanas 29-30:** Construir o "SPI simulado" — um serviço central que recebe mensagens de um banco origem e roteia pro banco destino. Idempotência ponta a ponta no roteamento.
- **Semanas 31-32:** Tratamento de falhas distribuídas. Padrão Saga pra transferências entre bancos. Compensação. Estados intermediários.
- **Semanas 33-34:** Observabilidade séria. Traces distribuídos atravessando múltiplos serviços. Dashboards de SLI/SLO.
- **Semanas 35-36:** Chaos testing — derrubar serviços aleatoriamente, ver o sistema se recuperar. Documentar comportamento. Relatório de Fase 3.

## Fase 4 (mês 10-12) — Inteligência + TCC

- **Semanas 37-38:** Definir escopo do componente de fraude. Estudar isolation forest, autoencoders, ou abordagens de grafo. Escolher e justificar.
- **Semanas 39-40:** Construir o serviço Python que consome o stream de eventos do RabbitMQ.
- **Semanas 41-42:** Treinar modelo com dados sintéticos que você gera (já que o sistema todo é simulado, você pode injetar fraudes propositalmente).
- **Semanas 43-44:** Integrar — quando uma transferência é detectada como suspeita, retornar à API que pode bloquear ou pedir confirmação extra.
- **Semanas 45-48:** Escrita do TCC. Tema sugerido: "Estratégias de garantia de idempotência e consistência em sistemas de pagamentos instantâneos: uma análise comparativa". Ou, se quiser viés IA: "Detecção de anomalias em redes de pagamentos instantâneos via [técnica escolhida]".

---

# Estudos paralelos — leituras e cursos

> Não vou te enterrar. Curadoria de poucos itens, escalonados ao longo dos 12 meses.

## Pra Fase 1 (mês 1-3)

1. **Designing Data-Intensive Applications, Martin Kleppmann** — capítulos 1-3 e 7. Esse livro vai te acompanhar o projeto inteiro. É denso, leia 10 páginas por semana, não tente devorar. **Prioridade máxima.**
2. **Use The Index, Luke** (use-the-index-luke.com) — gratuito, online. Ensina índices SQL como ninguém. Leia inteiro durante a semana 7.
3. **Documentação oficial do Go** + **Effective Go**. Você usa Go há 6 meses; agora é hora de ler a doc oficial inteira, não só Stack Overflow.
4. **The Twelve-Factor App** (12factor.net) — uma tarde de leitura. Vai influenciar toda configuração do projeto.

## Pra Fase 2 (mês 4-6)

5. DDIA capítulos 5, 8, 9, 11.
6. **"Patterns of Enterprise Application Architecture", Fowler** — só os capítulos sobre Unit of Work, Identity Map, e o post do Fowler sobre Outbox.
7. Documentação oficial de Postgres sobre `MVCC`, `isolation levels`, `locks`. Tá em inglês mas é incrível.

## Pra Fase 3 (mês 7-9)

8. DDIA capítulos 8, 9, 11 (releitura, agora faz mais sentido).
9. **"Site Reliability Engineering", Google** (gratuito online) — capítulos sobre SLI/SLO, postmortems, monitoramento.
10. Documentação OpenTelemetry.

## Pra Fase 4 (mês 10-12)

11. **"Hands-On Machine Learning", Géron** — capítulos sobre detecção de anomalia.
12. Papers acadêmicos específicos pro tema do TCC (busque no Google Scholar quando chegar lá).

---

# Riscos e como mitigar

- **Travar em uma semana e não conseguir avançar.** Mitigação: a regra é avançar ao final do dia 7 da semana, mesmo com débito técnico. O débito vai pra `BACKLOG.md` e é endereçado depois. Não é aceitável passar 3 semanas na mesma coisa.
- **Síndrome do tutorial.** Mitigação: regra de ouro do ritual diário (no outro arquivo).
- **Perfeccionismo.** Mitigação: ao final de cada semana, mesmo feio, commita e abre PR pra você mesmo. Merge depois de revisar friamente.
- **Vida acontece (faculdade pesa, estágio pesa).** Mitigação: política do "1 dia ruim é OK, 3 dias ruins é alarme". Se ficar 3 dias sem tocar, na próxima sessão você não tenta fazer o que perdeu — só retoma de onde estava.
- **Comparação com projetos de outros.** Mitigação: você está construindo *seu* sistema com profundidade. Tem pessoa com 10 projetos no GitHub, todos rasos. Você terá 1, profundo. Isso pesa mais.
