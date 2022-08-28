package sender

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gitlab.com/g6834/team31/auth/pkg/logging"
	"gitlab.com/g6834/team31/mailsender/internal/config"
	"gitlab.com/g6834/team31/mailsender/internal/domain/models"
)

type Sender struct {
	cfg    config.Sender
	queue  chan models.Email
	ticker *time.Ticker
	logger *logging.Logger
}

func New(cfg config.Sender, l *logging.Logger) *Sender {
	return &Sender{
		cfg:    cfg,
		queue:  make(chan models.Email, cfg.BuffSize),
		ticker: time.NewTicker(time.Second / time.Duration(cfg.RPS)),
		logger: l,
	}
}

func (s *Sender) AcceptMail(in models.Email) {
	s.queue <- in
}

var once sync.Once

func (s *Sender) SendMail() error {
	mail := <-s.queue
	if a := rand.Intn(10); a > 8 {
		return fmt.Errorf("couldn't send email %+v", mail)
	}
	fmt.Printf("mail sent! %+v\n\n\n", mail)
	return nil
}

func (s *Sender) LimitedSend(ctx context.Context) chan error {
	errChan := make(chan error)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				err := s.SendMail()
				if err != nil {
					errChan <- err
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return errChan
}
