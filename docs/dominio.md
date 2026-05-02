# Como funciona o SPI, o DICT, mensagens ISO 20022

# 1. Visão geral do ecossistema PIX

```
┌─────────────────────────────────────────────────────────────────┐
│                         BACEN (Banco Central)                   │
│                                                                 │
│    ┌──────────────┐          ┌──────────────────────────────┐   │
│    │     SPI      │          │           DICT               │   │
│    │  (liquidação)│          │  (diretório de chaves/contas)│   │
│    └──────────────┘          └──────────────────────────────┘   │
└────────────┬────────────────────────────┬───────────────────────┘
             │ mensagens ISO 20022         │ REST API
             │ via HTTPS mTLS             │ via HTTPS mTLS
    ┌────────┴────────┐         ┌─────────┴────────┐
    │    Banco A      │         │     Banco B       │
    │  (Participante) │         │   (Participante)  │
    └─────────────────┘         └──────────────────┘
```

Os três são independentes mas se integram:

- **DICT** → "onde está essa chave Pix? qual conta/banco?"
- **SPI** → "liquida o dinheiro entre os bancos"
- **ISO 20022** → "qual é o formato das mensagens trocadas"

## 1.1 SPI — Sistema de Pagamentos Instantâneos

O SPI é a câmara de compensação e liquidação do Pix. Ele processa transferências entre participantes (bancos, fintechs, etc.) em tempo real, 24/7.

**Fluxo:**

```
Pagador (João) → Banco A → SPI → Banco B → Beneficiário (Maria)
```

**Passo a passo detalhado:**

1. João (cliente do Banco A) inicia transferência de R$100 para Maria
   - Banco A valida: João tem saldo? Conta ativa? Limites?

2. Banco A envia mensagem pacs.008 ao SPI
   - "Quero transferir R$100 do Banco A para o Banco B"
   - Assinado digitalmente com certificado ICP-Brasil

3. SPI processa:
   - Valida a mensagem (schema, assinatura digital)
   - Verifica se Banco A tem saldo na conta Reservas no BACEN
   - Se sim: debita Banco A, credita Banco B (atomicamente)
   - Envia confirmação pacs.002 para o Banco A

4. SPI entrega a instrução de crédito ao Banco B (pacs.008)
   - Banco B tem até 10s para creditar Maria

5. Banco B responde com pacs.002 confirmando crédito
   - Se não responder em 10s → pacs.002 com erro → reversão

> **Tempo total:** < 10 segundos (regulatório)

### Conta Reservas (o mecanismo de liquidação)

```
┌─────────────────────────────────────────────┐
│             BACEN — Conta Reservas          │
│                                             │
│  Banco A: R$ 50.000.000 (saldo)            │
│  Banco B: R$ 30.000.000 (saldo)            │
│  Banco C: R$ 10.000.000 (saldo)            │
└─────────────────────────────────────────────┘
```

Transferência de R$100 do Banco A → Banco B:

- Banco A: R$50.000.000 → R$49.999.900
- Banco B: R$30.000.000 → R$30.000.100

A liquidação é **bruta e em tempo real (LBTR)**. Cada transação liquida individualmente, sem compensação em lote.

## 1.2 DICT — Diretório de Identificadores de Contas Transacionais

O DICT é um banco de dados centralizado no BACEN que mapeia "chave Pix" → "conta bancária".

**Tipos de chave e o que o DICT armazena:**

```json
{
  "key": "joao@email.com",
  "keyType": "EMAIL",
  "account": {
    "participant": "60701190",
    "branch": "0001",
    "accountNumber": "123456",
    "accountType": "CACC",
    "openingDate": "2021-01-15"
  },
  "owner": {
    "name": "João da Silva",
    "taxIdNumber": "***123456**",
    "type": "NATURAL_PERSON"
  },
  "keyCreationDate": "2021-03-10T14:22:00Z",
  "keyOwnershipDate": "2021-03-10T14:22:00Z"
}
```

### 1.2.1. Consulta ao DICT

Quando o usuário digita uma chave Pix, o banco pagador consulta o DICT antes de qualquer transferência:

```
GET /api/v1/directory/keys/joao@email.com
```

```json
{
  "key": "joao@email.com",
  "account": {
    "participant": "60701190",
    "branch": "0001",
    "accountNumber": "123456",
    "accountType": "CACC"
  },
  "owner": {
    "name": "João da Silva"
  }
}
```

O Banco A usa isso para saber para qual banco e conta enviar a mensagem pacs.008.

### 1.2.2. Operações do DICT

| Operação | Descrição |
|---|---|
| `POST /keys` | Cadastrar chave Pix |
| `DELETE /keys/{key}` | Excluir chave |
| `GET /keys/{key}` | Consultar chave (quem paga) |
| `POST /claims` | Portabilidade / reivindicação de chave |
| `POST /infringementReports` | Reportar fraude |

## 1.3. ISO 20022 — O formato das mensagens

ISO 20022 é um padrão internacional de mensagens financeiras baseado em XML. O Pix usa um subconjunto específico chamado "Pix Payments".

### 1.3.1 Mensagens principais

| Mensagem | Nome | Quando é usada |
|---|---|---|
| `pacs.008` | FIToFICustomerCreditTransfer | Banco A → SPI → Banco B (a transferência em si) |
| `pacs.002` | FIToFIPaymentStatusReport | Confirmação ou rejeição da transferência |
| `pacs.004` | PaymentReturn | Devolução de uma transferência |
| `pain.001` | CustomerCreditTransferInitiation | Cliente → Banco (inicia a transferência) |
| `pain.002` | CustomerPaymentStatusReport | Status do pain.001 |

### 1.3.2. pacs.008 — exemplo real

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.008.001.08">
  <FIToFICstmrCdtTrf>

    <!-- Cabeçalho do grupo -->
    <GrpHdr>
      <MsgId>MSG-BANCO-A-20240101-00001</MsgId>
      <CreDtTm>2024-01-01T10:30:00.000Z</CreDtTm>
      <NbOfTxs>1</NbOfTxs>
      <TtlIntrBkSttlmAmt Ccy="BRL">100.00</TtlIntrBkSttlmAmt>
      <IntrBkSttlmDt>2024-01-01</IntrBkSttlmDt>
      <SttlmInf>
        <SttlmMtd>CLRG</SttlmMtd>
      </SttlmInf>
    </GrpHdr>

    <!-- A transação em si -->
    <CdtTrfTxInf>

      <!-- ID de endToEnd — chave de idempotência do Pix! -->
      <PmtId>
        <EndToEndId>E60746948202401011030ABC123456</EndToEndId>
        <!-- Formato: E + ISPB_pagador + AAAAMMDD + HHMM + sufixo_11_chars -->
      </PmtId>

      <IntrBkSttlmAmt Ccy="BRL">100.00</IntrBkSttlmAmt>

      <!-- Banco pagador (Banco A) -->
      <InstgAgt>
        <FinInstnId>
          <ClrSysMmbId>
            <MmbId>60746948</MmbId>
          </ClrSysMmbId>
        </FinInstnId>
      </InstgAgt>

      <!-- Banco recebedor (Banco B) -->
      <InstdAgt>
        <FinInstnId>
          <ClrSysMmbId>
            <MmbId>60701190</MmbId>
          </ClrSysMmbId>
        </FinInstnId>
      </InstdAgt>

      <!-- Pagador (João) -->
      <Dbtr>
        <Nm>João da Silva</Nm>
        <Id>
          <PrvtId>
            <TaxId>12345678901</TaxId>
          </PrvtId>
        </Id>
      </Dbtr>
      <DbtrAcct>
        <Id>
          <Othr>
            <Id>000010001234567</Id>
          </Othr>
        </Id>
        <Tp><Cd>CACC</Cd></Tp>
      </DbtrAcct>

      <!-- Beneficiário (Maria) -->
      <Cdtr>
        <Nm>Maria Santos</Nm>
        <Id>
          <PrvtId>
            <TaxId>98765432100</TaxId>
          </PrvtId>
        </Id>
      </Cdtr>
      <CdtrAcct>
        <Id>
          <Othr>
            <Id>000010007654321</Id>
          </Othr>
        </Id>
        <Tp><Cd>CACC</Cd></Tp>
      </CdtrAcct>

      <RmtInf>
        <Ustrd>Pagamento almoço</Ustrd>
      </RmtInf>

      <SplmtryData>
        <Envlp>
          <Document>
            <TxInfAndSts>
              <TxSts>ACSC</TxSts>
            </TxInfAndSts>
          </Document>
        </Envlp>
      </SplmtryData>

    </CdtTrfTxInf>
  </FIToFICstmrCdtTrf>
</Document>
```

#### 1.3.2.1. Exemplo: pacs.002 (confirmação)

```xml
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pacs.002.001.10">
  <FIToFIPmtStsRpt>
    <GrpHdr>
      <MsgId>STATUS-BANCO-B-20240101-00001</MsgId>
      <CreDtTm>2024-01-01T10:30:05.123Z</CreDtTm>
    </GrpHdr>
    <TxInfAndSts>
      <OrgnlEndToEndId>E60746948202401011030ABC123456</OrgnlEndToEndId>
      <TxSts>ACSC</TxSts>
      <!-- Se rejeitado:
      <StsRsnInf>
        <Rsn><Cd>AC01</Cd></Rsn>
      </StsRsnInf> -->
    </TxInfAndSts>
  </FIToFIPmtStsRpt>
</Document>
```

#### 1.2.3.2. Códigos de status importantes

| Código | Significado |
|---|---|
| `ACSC` | Aceito e liquidado (sucesso) |
| `ACSP` | Aceito, processamento em andamento |
| `RJCT` | Rejeitado |
| `AC01` | Conta destino inválida |
| `AC03` | Conta destino encerrada |
| `AM04` | Saldo insuficiente |
| `DS04` | Erro de assinatura digital |
| `ED05` | Liquidação falhou |

## 1.3. O EndToEndId — a chave de idempotência nativa do Pix

Campo fundamental para o projeto: mecanismo de idempotência do Pix.

```
E  60746948  20240101  1030  ABC12345678
│  │         │         │     │
│  │         │         │     └── 11 chars aleatórios (sufixo único)
│  │         │         └──────── HHMM (hora/minuto UTC)
│  │         └────────────────── AAAAMMDD (data)
│  └──────────────────────────── ISPB do participante pagador (8 dígitos)
└─────────────────────────────── Literal "E"
```

**Regras:**

- Cada `EndToEndId` deve ser único por participante
- O SPI rejeita duplicatas: `RJCT` com código específico
- Serve para rastrear a transação de ponta a ponta entre bancos
- O projeto simulado deve implementar isso — é a chave de idempotência da Fase 2

## 1.4. Conceitos importantes aprendidos hoje

| Conceito | Por que importa para o projeto_pay |
|---|---|
| `EndToEndId` | Chave de idempotência (Fase 2, Semana 13) |
| LBTR | Justifica por que liquidação deve ser atômica no banco |
| DICT | Simulado com uma tabela `pix_keys` |
| `pacs.008` | Estrutura que o JSON de transferência vai espelhar |
| SLA 10s | Meta de latência para os testes de carga (k6) |
| Participant Code (ISPB) | Cada "banco" simulado na Fase 3 tem um código único |
