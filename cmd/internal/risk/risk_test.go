package risk

import (
	"testing"
	"time"
)

// horário comercial em São Paulo
func daytime(cfg Config) time.Time {
	return time.Date(2026, 10, 3, 14, 0, 0, 0, cfg.Location)
}

func night(cfg Config) time.Time {
	return time.Date(2026, 10, 3, 23, 0, 0, 0, cfg.Location)
}

// sinais de um cliente comum mandando para um contato conhecido
func normalSignals(cfg Config) Signals {
	return Signals{
		Amount:             5_000,
		SenderBalance:      100_000,
		Now:                daytime(cfg),
		TransfersLast10Min: 0,
		HistoryCount:       10,
		AvgAmount:          4_000,
		FirstTimeReceiver:  false,
		ReceiverAccountAge: 365 * 24 * time.Hour,
	}
}

func TestEvaluate_ApprovesNormalTransfer(t *testing.T) {
	cfg := DefaultConfig()
	a := Evaluate(cfg, normalSignals(cfg))

	if a.Decision != DecisionApprove {
		t.Fatalf("esperava APPROVE, veio %s (%v)", a.Decision, a.RedFlags)
	}
	if a.FraudScore != 0 || a.FraudLevel != LevelLow {
		t.Fatalf("esperava score 0/LOW, veio %d/%s", a.FraudScore, a.FraudLevel)
	}
	if a.Explanation == "" {
		t.Fatal("toda decisão precisa de explicação")
	}
}

func TestEvaluate_NightLimitBlocksRegardlessOfScore(t *testing.T) {
	cfg := DefaultConfig()
	s := normalSignals(cfg)
	s.Now = night(cfg)
	s.Amount = cfg.NightLimit + 1
	s.AvgAmount = s.Amount // não dispara o sinal de valor atípico

	a := Evaluate(cfg, s)
	if a.Decision != DecisionBlock {
		t.Fatalf("esperava BLOCK, veio %s", a.Decision)
	}
	if len(a.Compliance) != 1 {
		t.Fatalf("esperava 1 violação, veio %v", a.Compliance)
	}
}

func TestEvaluate_NightLimitAllowsExactLimit(t *testing.T) {
	cfg := DefaultConfig()
	s := normalSignals(cfg)
	s.Now = night(cfg)
	s.Amount = cfg.NightLimit
	s.AvgAmount = s.Amount

	if a := Evaluate(cfg, s); a.Decision != DecisionApprove {
		t.Fatalf("esperava APPROVE no limite exato, veio %s (%v %v)", a.Decision, a.Compliance, a.RedFlags)
	}
}

func TestEvaluate_HighRiskGoesToHumanReview(t *testing.T) {
	cfg := DefaultConfig()
	s := normalSignals(cfg)
	s.FirstTimeReceiver = true               // 15
	s.ReceiverAccountAge = time.Hour         // 20
	s.Amount = s.AvgAmount * 6               // 25
	s.SenderBalance = s.Amount + s.Amount/10 // 15 (esvazia a conta)

	a := Evaluate(cfg, s)
	if a.FraudScore != 75 || a.FraudLevel != LevelHigh {
		t.Fatalf("esperava 75/HIGH, veio %d/%s (%v)", a.FraudScore, a.FraudLevel, a.RedFlags)
	}
	if a.Decision != DecisionReview {
		t.Fatalf("esperava REVIEW, veio %s", a.Decision)
	}
	if len(a.RedFlags) != 4 {
		t.Fatalf("esperava 4 red flags, veio %v", a.RedFlags)
	}
}

func TestEvaluate_CriticalRiskIsVetoed(t *testing.T) {
	cfg := DefaultConfig()
	s := normalSignals(cfg)
	s.TransfersLast10Min = cfg.VelocityThreshold // 35
	s.FirstTimeReceiver = true                   // 15
	s.ReceiverAccountAge = time.Hour             // 20
	s.Amount = s.AvgAmount * 6                   // 25

	a := Evaluate(cfg, s)
	if a.FraudLevel != LevelCritical || a.Decision != DecisionBlock {
		t.Fatalf("esperava CRITICAL/BLOCK, veio %s/%s (score %d)", a.FraudLevel, a.Decision, a.FraudScore)
	}
}

func TestFraudAgent_ScoreIsCappedAt100(t *testing.T) {
	cfg := DefaultConfig()
	s := Signals{
		Amount:             1_000_000,
		SenderBalance:      1_000_000,
		Now:                night(cfg),
		TransfersLast10Min: 10,
		HistoryCount:       10,
		AvgAmount:          1_000,
		FirstTimeReceiver:  true,
		ReceiverAccountAge: 0,
	}
	if score, _ := fraudAgent(cfg, s); score != 100 {
		t.Fatalf("esperava 100, veio %d", score)
	}
}

func TestFraudAgent_SpikeNeedsHistory(t *testing.T) {
	cfg := DefaultConfig()
	s := normalSignals(cfg)
	s.HistoryCount = cfg.MinHistoryForSpike - 1
	s.Amount = s.AvgAmount * 10
	s.SenderBalance = s.Amount * 10

	if score, flags := fraudAgent(cfg, s); score != 0 {
		t.Fatalf("sem histórico não deveria pontuar valor atípico, veio %d %v", score, flags)
	}
}

func TestIsNight(t *testing.T) {
	cfg := DefaultConfig()
	cases := []struct {
		hour int
		want bool
	}{
		{5, true}, {6, false}, {12, false}, {19, false}, {20, true}, {23, true}, {0, true},
	}
	for _, c := range cases {
		got := isNight(cfg, time.Date(2026, 10, 3, c.hour, 0, 0, 0, cfg.Location))
		if got != c.want {
			t.Errorf("%dh: esperava %v, veio %v", c.hour, c.want, got)
		}
	}
}
