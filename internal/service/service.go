package service

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SayHello() string {
	return "Hello, World!"
}
