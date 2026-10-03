# Análise de risco das transferências

Toda chamada a `POST /transfers` passa por uma avaliação de risco antes de mexer em saldo.
O desenho vem da palestra "Agentes financeiros" (TDC São Paulo 2026), que propõe agentes
especialistas coordenados por um orquestrador. Aqui não há ML nem LLM: os agentes são regras
determinísticas, o que mantém a decisão auditável e rápida.

| Agente | O que faz | Pode |
|---|---|---|
| Fraude | Soma sinais num score 0-100 e classifica em LOW, MEDIUM, HIGH ou CRITICAL | Vetar (CRITICAL) ou mandar para revisão (HIGH) |
| Compliance | Checa regras regulatórias; hoje, o limite noturno de R$ 1.000 entre 20h e 6h | Bloquear, sem exceções |
| Orquestrador (`risk.Evaluate`) | Junta os dois; compliance tem precedência | Decidir APPROVE, REVIEW ou BLOCK |

## Sinais do agente de fraude

| Sinal | Pontos |
|---|---|
| 5 ou mais transferências do remetente nos últimos 10 minutos | 35 |
| Valor mais de 5x a média dos últimos 30 dias (com ao menos 3 transferências) | 25 |
| Conta recebedora criada há menos de 24h | 20 |
| Primeira transferência para este recebedor | 15 |
| Transferência leva 80% ou mais do saldo | 15 |
| Recebedor novo no período noturno | 10 |

Faixas: 0-29 LOW, 30-59 MEDIUM, 60-79 HIGH, 80-100 CRITICAL. Limites ficam em `risk.DefaultConfig`.

## Resposta da API

APPROVE segue o fluxo normal (201). REVIEW e BLOCK não executam a transferência e respondem
422 com a avaliação completa:

```json
{
  "decision": "REVIEW",
  "fraudScore": 75,
  "fraudLevel": "HIGH",
  "redFlags": ["primeira transferência para este recebedor", "conta recebedora criada há menos de 24h"],
  "complianceViolations": [],
  "explanation": "Transferência retida para revisão humana (risco HIGH)."
}
```

## Trilha de auditoria

Cada avaliação vira uma linha em `risk_decisions` (migração 000007), com score, nível, motivos
e o `transfer_id` quando a transferência foi executada.

## Próximos passos

- Fila de revisão: endpoint para um analista aprovar ou recusar as transferências em REVIEW.
- Limites por cliente em vez dos padrões globais.
- Métricas por agente (taxa de REVIEW e BLOCK) para perceber regra calibrada errado.
