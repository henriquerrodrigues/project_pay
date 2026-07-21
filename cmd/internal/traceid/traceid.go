package traceid

import "context"

type ctxKey struct{}

// NewContext retorna um context carregando o trace ID.
func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext retorna o trace ID do context, ou "" se ausente.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}
