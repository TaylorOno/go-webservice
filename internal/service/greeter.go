package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/taylorono/go-lib/traces"
)

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

func (s *Greeter) SayHello(ctx context.Context) string {
	_, span := traces.Start(ctx, "SayHello", traces.AsComponent("greeter"))
	defer func() { span.End() }()

	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	return "Hello, World!"
}

func (s *Greeter) SayHelloUser(ctx context.Context, name string) string {
	_, span := traces.Start(ctx, "SayHelloUser", traces.AsComponent("greeter"))
	defer func() { span.End() }()

	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	return "Hello, " + name + "!"
}

func (s *Greeter) SayMorningJokes(ctx context.Context) (string, error) {
	ctx, span := traces.Start(ctx, "SayMorningJokes", traces.AsComponent("greeter"))
	defer func() { span.End() }()

	s.joker.GetJoke(ctx)

	_, sleepSpan := traces.Start(ctx, "Sleeping", traces.AsComponent("greeter"))
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	sleepSpan.End()

	return s.joker.GetJoke(ctx)
}
