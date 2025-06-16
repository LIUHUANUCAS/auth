package main

import (
	"context"
	"net"

	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"

	// "golang.ngrok.com/ngrok/config"
	authConfig "github.com/LIUHUANUCAS/auth/config"
)

func newNgrokListener(ctx context.Context, cfg *authConfig.Config) (net.Listener, error) {
	return ngrok.Listen(ctx,
		config.HTTPEndpoint(
			config.WithURL(cfg.Ngrok.HostName),
		),
		ngrok.WithAuthtokenFromEnv(),
	)
}
