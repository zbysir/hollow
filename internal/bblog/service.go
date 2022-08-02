package bblog

import (
	"context"
	"net/http"
)

// Service 本地运行的服务，用于实时预览
type Service struct {
	s  *http.Server
	sm *http.ServeMux
}

func NewService(addr string) (*Service, error) {
	x := http.NewServeMux()

	httpServer := &http.Server{
		Addr:    addr,
		Handler: x,
	}

	return &Service{s: httpServer, sm: x}, nil
}

func (s *Service) Handler(path string, f func(writer http.ResponseWriter, request *http.Request)) {
	s.sm.HandleFunc(path, f)
}

func (s Service) Start(ctx context.Context) error {
	var err error
	go func() {
		select {
		case <-ctx.Done():
			err = s.s.Shutdown(ctx)
			if err != nil {
				return
			}
		}
	}()
	err = s.s.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}
