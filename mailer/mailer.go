package mailer

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

type Email struct {
	From    string
	To      string
	Subject string
	Text    string
	HTML    string
}

type Provider interface {
	Send(ctx context.Context, email Email) error
	Name() string
}

type Mailer struct {
	providers []Provider
}

func New(providers ...Provider) *Mailer {
	return &Mailer{providers: providers}
}

type SendOptions struct {
	MaxAttempts int
	RetryDelay  time.Duration
}

type Option func(*SendOptions)

func WithMaxAttempts(attempts int) Option {
	return func(o *SendOptions) {
		if attempts > 0 {
			o.MaxAttempts = attempts
		}
	}
}

func WithRetryDelay(delay time.Duration) Option {
	return func(o *SendOptions) {
		if delay >= 0 {
			o.RetryDelay = delay
		}
	}
}

func (m *Mailer) Send(ctx context.Context, email Email, opts ...Option) error {
	if len(m.providers) == 0 {
		return errors.New("no email providers configured")
	}

	options := SendOptions{
		MaxAttempts: len(m.providers),
		RetryDelay:  3 * time.Second,
	}

	for _, optFunc := range opts {
		optFunc(&options)
	}

	providerIndices := make([]int, len(m.providers))
	for i := range providerIndices {
		providerIndices[i] = i
	}
	rand.Shuffle(len(providerIndices), func(i, j int) {
		providerIndices[i], providerIndices[j] = providerIndices[j], providerIndices[i]
	})

	var lastErr error
	attemptsMade := 0

	for _, idx := range providerIndices {
		select {
		case <-ctx.Done():
			return fmt.Errorf("email send canceled: %w", ctx.Err())
		default:
		}

		if attemptsMade >= options.MaxAttempts {
			break
		}

		provider := m.providers[idx]
		attemptsMade++

		err := provider.Send(ctx, email)
		if err == nil {
			return nil
		}

		lastErr = fmt.Errorf("provider %s failed on attempt %d: %w", provider.Name(), attemptsMade, err)

		if attemptsMade < options.MaxAttempts && options.RetryDelay > 0 {
			timer := time.NewTimer(options.RetryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return fmt.Errorf("email send canceled during retry delay: %w", ctx.Err())
			case <-timer.C:
			}
		}
	}

	if lastErr != nil {
		return fmt.Errorf("all email sending attempts failed after %d tries: %w", attemptsMade, lastErr)
	}

	return errors.New("no email providers available or all attempts failed without an explicit error")
}
