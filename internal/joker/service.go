package joker

import (
	"context"
	"fmt"
	"net/http"

	"github.com/taylorono/go-lib/rest"
	"github.com/taylorono/go-webservice/internal/service"
)

var _ service.JokeProvider = &Service{}

type Service struct {
	client *rest.Client
}

func NewJokeProvider(client *rest.Client) *Service {
	return &Service{client: client}
}

type Joke struct {
	Type      string `json:"type"`
	Setup     string `json:"setup"`
	Punchline string `json:"punchline"`
	Id        int    `json:"id"`
}

func (s Service) GetJoke(ctx context.Context) (string, error) {
	req, err := http.NewRequest("GET", "https://official-joke-api.appspot.com/random_joke", nil)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}

	joke, err := rest.Decode[Joke](resp)
	if err != nil {
		return "", fmt.Errorf("failed to decode joke: %w", err)
	}

	return fmt.Sprintf("Q: %s\nA: %s ", joke.Setup, joke.Punchline), nil
}
