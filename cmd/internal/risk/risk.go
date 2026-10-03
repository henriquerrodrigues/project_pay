// Package risk avalia uma transferência antes de ela ser executada.
//
// O desenho segue a arquitetura multi-agentes da palestra "Agentes financeiros"
// (TDC São Paulo 2026), adaptada para um motor PIX sem ML:
//
//   - agente de fraude: soma sinais de risco num score 0-100 e classifica em
//     LOW/MEDIUM/HIGH/CRITICAL; pode vetar a transferência;
//   - agente de compliance: regras regulatórias que BLOQUEIAM se qualquer
//     checagem falhar, sem exceções;
//   - orquestrador (Evaluate): combina os dois e decide APPROVE, REVIEW
//     (humano no loop) ou BLOCK;
//   - explicabilidade: toda decisão carrega os motivos em português.
//
// O pacote é puro (sem banco, sem relógio): quem chama monta os Signals.
package risk

import "time"

type Level string

const (
	LevelLow      Level = "LOW"
	LevelMedium   Level = "MEDIUM"
	LevelHigh     Level = "HIGH"
	LevelCritical Level = "CRITICAL"
)

type Decision string

const (
	DecisionApprove Decision = "APPROVE"
	DecisionReview  Decision = "REVIEW"
	DecisionBlock   Decision = "BLOCK"
)

// Signals são os dados que os agentes analisam. Valores em centavos.
type Signals struct {
	Amount        int64
	SenderBalance int64
	Now           time.Time

	// histórico do remetente
	TransfersLast10Min int
	HistoryCount       int   // transferências nos últimos 30 dias
	AvgAmount          int64 // média dos últimos 30 dias

	// relação remetente -> recebedor
	FirstTimeReceiver  bool
	ReceiverAccountAge time.Duration
}

// Config guarda os limites. Use DefaultConfig e ajuste o que precisar.
type Config struct {
	Location *time.Location

	// compliance: limite noturno (BACEN, Res. 142/2021) para pessoa física
	NightStartHour int
	NightEndHour   int
	NightLimit     int64

	// fraude
	VelocityThreshold   int     // transferências em 10 minutos
	AmountSpikeFactor   float64 // valor > fator * média
	MinHistoryForSpike  int
	DrainRatio          float64 // valor >= ratio * saldo
	NewAccountThreshold time.Duration

	ReviewFrom Level // a partir deste nível vai para revisão humana
	BlockFrom  Level // a partir deste nível o agente de fraude veta
}

func DefaultConfig() Config {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*60*60)
	}
	return Config{
		Location:            loc,
		NightStartHour:      20,
		NightEndHour:        6,
		NightLimit:          100_000, // R$ 1.000,00
		VelocityThreshold:   5,
		AmountSpikeFactor:   5,
		MinHistoryForSpike:  3,
		DrainRatio:          0.8,
		NewAccountThreshold: 24 * time.Hour,
		ReviewFrom:          LevelHigh,
		BlockFrom:           LevelCritical,
	}
}

// Assessment é a resposta do orquestrador, pronta para auditoria.
type Assessment struct {
	Decision    Decision `json:"decision"`
	FraudScore  int      `json:"fraudScore"`
	FraudLevel  Level    `json:"fraudLevel"`
	RedFlags    []string `json:"redFlags"`
	Compliance  []string `json:"complianceViolations"`
	Explanation string   `json:"explanation"`
}

// Evaluate roda os agentes e decide. Compliance tem precedência sobre fraude.
func Evaluate(cfg Config, s Signals) Assessment {
	score, flags := fraudAgent(cfg, s)
	level := levelFor(score)
	violations := complianceAgent(cfg, s)

	a := Assessment{
		FraudScore: score,
		FraudLevel: level,
		RedFlags:   flags,
		Compliance: violations,
	}

	switch {
	case len(violations) > 0:
		a.Decision = DecisionBlock
		a.Explanation = "Transferência bloqueada por regra regulatória: " + violations[0] + "."
	case rank(level) >= rank(cfg.BlockFrom):
		a.Decision = DecisionBlock
		a.Explanation = "Transferência bloqueada pelo antifraude (risco " + string(level) + ")."
	case rank(level) >= rank(cfg.ReviewFrom):
		a.Decision = DecisionReview
		a.Explanation = "Transferência retida para revisão humana (risco " + string(level) + ")."
	default:
		a.Decision = DecisionApprove
		a.Explanation = "Transferência aprovada (risco " + string(level) + ")."
	}
	return a
}

// fraudAgent soma pesos de cada sinal e devolve o score (0-100) e os motivos.
func fraudAgent(cfg Config, s Signals) (int, []string) {
	score := 0
	flags := []string{}
	add := func(points int, flag string) {
		score += points
		flags = append(flags, flag)
	}

	if s.TransfersLast10Min >= cfg.VelocityThreshold {
		add(35, "muitas transferências em poucos minutos")
	}
	if s.FirstTimeReceiver {
		add(15, "primeira transferência para este recebedor")
	}
	if s.HistoryCount >= cfg.MinHistoryForSpike && s.AvgAmount > 0 &&
		float64(s.Amount) > cfg.AmountSpikeFactor*float64(s.AvgAmount) {
		add(25, "valor muito acima da média do remetente")
	}
	if s.SenderBalance > 0 && float64(s.Amount) >= cfg.DrainRatio*float64(s.SenderBalance) {
		add(15, "transferência esvazia a maior parte do saldo")
	}
	if s.ReceiverAccountAge < cfg.NewAccountThreshold {
		add(20, "conta recebedora criada há menos de 24h")
	}
	if isNight(cfg, s.Now) && s.FirstTimeReceiver {
		add(10, "recebedor novo no período noturno")
	}

	if score > 100 {
		score = 100
	}
	return score, flags
}

// complianceAgent devolve as regras violadas. Qualquer violação bloqueia.
func complianceAgent(cfg Config, s Signals) []string {
	violations := []string{}
	if isNight(cfg, s.Now) && s.Amount > cfg.NightLimit {
		violations = append(violations, "valor acima do limite noturno (20h às 6h)")
	}
	return violations
}

func isNight(cfg Config, t time.Time) bool {
	h := t.In(cfg.Location).Hour()
	return h >= cfg.NightStartHour || h < cfg.NightEndHour
}

func levelFor(score int) Level {
	switch {
	case score >= 80:
		return LevelCritical
	case score >= 60:
		return LevelHigh
	case score >= 30:
		return LevelMedium
	default:
		return LevelLow
	}
}

func rank(l Level) int {
	switch l {
	case LevelMedium:
		return 1
	case LevelHigh:
		return 2
	case LevelCritical:
		return 3
	default:
		return 0
	}
}
