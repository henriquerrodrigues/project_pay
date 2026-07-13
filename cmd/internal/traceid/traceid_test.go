package traceid_test

import (
	"context"
	"testing"

	"project_pay/cmd/internal/traceid"
)

func TestRoundTrip(t *testing.T) {
	ctx := traceid.NewContext(context.Background(), "abc-123")
	if got := traceid.FromContext(ctx); got != "abc-123" {
		t.Errorf("FromContext = %q, esperava %q", got, "abc-123")
	}
}

func TestMissing(t *testing.T) {
	if got := traceid.FromContext(context.Background()); got != "" {
		t.Errorf("FromContext em context vazio = %q, esperava vazio", got)
	}
}
