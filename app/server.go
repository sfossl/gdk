package app

import (
	stdctx "context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/sfossl/gda/interaction"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
)

func (a *App) runGateway(ctx stdctx.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Gateway.Connect(ctx)
	}()

	select {
	case <-ctx.Done():
		_ = a.Gateway.Close()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (a *App) webhookHandler() (http.Handler, error) {
	if a.config.PublicKey == "" {
		return nil, ErrPublicKeyRequired
	}

	pubKey, err := interaction.ParsePublicKey(a.config.PublicKey)
	if err != nil {
		return nil, ErrInvalidSignature
	}

	mux := http.NewServeMux()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var in types.Interaction
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		bgCtx, cancel := stdctx.WithTimeout(stdctx.WithoutCancel(r.Context()), a.config.Timeout)
		gdkCtx := context.New(bgCtx, &in, a.REST)

		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					cancel()
				}
			}()
			err := a.Handle(gdkCtx)
			if err != nil {
				cancel()
			}
		}()

		timer := time.NewTimer(2800 * time.Millisecond)
		defer timer.Stop()

		select {
		case resp := <-gdkCtx.ResponseChannel():
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case <-timer.C:
			select {
			case resp := <-gdkCtx.ResponseChannel():
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			default:
				if isDeferrableInteraction(in.Type) {
					_ = gdkCtx.Defer(false)
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(types.InteractionResponse{
						Type: types.CallbackTypeDeferredChannelMessageWithSource,
					})
				} else if in.Type == types.InteractionTypeApplicationCommandAutocomplete {
					_ = gdkCtx.Autocomplete(nil)
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(types.InteractionResponse{
						Type: types.CallbackTypeApplicationCommandAutocompleteResult,
						Data: &types.InteractionResponseData{Choices: []types.Choice{}},
					})
				} else if in.Type == types.InteractionTypePing {
					_ = gdkCtx.Pong()
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(types.InteractionResponse{
						Type: types.CallbackTypePong,
					})
				}
			}
		case <-r.Context().Done():
			return
		}
	})

	verify := interaction.VerifyMiddleware(pubKey)
	mux.Handle(a.config.WebhookPath, verify(handler))
	return mux, nil
}

func (a *App) runWebhook(ctx stdctx.Context) error {
	handler, err := a.webhookHandler()
	if err != nil {
		return err
	}

	a.server = &http.Server{
		Addr:    a.config.WebhookAddr,
		Handler: handler,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := stdctx.WithTimeout(stdctx.Background(), 5*time.Second)
		defer cancel()
		return a.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (a *App) Close() error {
	if a.server != nil {
		ctx, cancel := stdctx.WithTimeout(stdctx.Background(), 5*time.Second)
		defer cancel()
		_ = a.server.Shutdown(ctx)
	}
	if a.Gateway != nil {
		_ = a.Gateway.Close()
	}
	return nil
}

