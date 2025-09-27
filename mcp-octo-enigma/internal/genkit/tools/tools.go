package tools

import "context"

type Tool func(ctx context.Context, args map[string]any) (any, error)
