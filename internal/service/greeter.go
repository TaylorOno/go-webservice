package service

import "context"

type JokeProvider interface {
	GetJoke(ctx context.Context) (string, error)
}

type Greeter struct {
	joker JokeProvider
}

func NewGreater(jokeService JokeProvider) *Greeter {
	return &Greeter{
		joker: jokeService,
	}
}

func (s *Greeter) SayHello() string {
	return "Hello, World!"
}

func (s *Greeter) SayHelloUser(name string) string {
	return "Hello, " + name + "!"
}

func (s *Greeter) SayMorningJokes(ctx context.Context) (string, error) {
	return s.joker.GetJoke(ctx)
}
