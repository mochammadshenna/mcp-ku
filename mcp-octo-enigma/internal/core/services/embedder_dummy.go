package services

import "context"

type DummyEmbedder struct{}

func (DummyEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	_ = ctx
	// return a small fixed-size vector to allow demo without external models
	return make([]float32, 1536), nil
}
