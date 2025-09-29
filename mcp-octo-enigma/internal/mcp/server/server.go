package server

import "context"

type Server struct{}

func New() *Server { return &Server{} }
func (s *Server) Start(ctx context.Context) error { _ = ctx; return nil }
