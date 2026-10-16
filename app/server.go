package app

import (
	stdctx "context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

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

func (a *App) runWebhook(ctx stdctx.Context) error {
	if a.config.PublicKey == "" {
		return ErrPublicKeyRequired
	}

	pubKeyBytes, err := hex.DecodeString(a.config.PublicKey)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return ErrInvalidSignature
	}

	mux := http.NewServeMux()
	mux.HandleFunc(a.config.WebhookPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sigHex := r.Header.Get("X-Signature-Ed25519")
		timestamp := r.Header.Get("X-Signature-Timestamp")
		if sigHex == "" || timestamp == "" {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		sec, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}
		diff := time.Since(time.Unix(sec, 0))
		if diff < -5*time.Minute || diff > 5*time.Minute {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		sig, err := hex.DecodeString(sigHex)
		if err != nil || len(sig) != ed25519.SignatureSize {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		msg := append([]byte(timestamp), body...)
		if !ed25519.Verify(pubKeyBytes, msg, sig) {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}

		var in types.Interaction
		if err := json.Unmarshal(body, &in); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		bgCtx, cancel := stdctx.WithTimeout(stdctx.WithoutCancel(r.Context()), a.config.Timeout)
		gdkCtx := context.New(bgCtx, &in, a.REST)

		go func() {
			defer cancel()
			_ = a.Handle(gdkCtx)
		}()

		deferDelay := a.config.AutoDeferDelay
		if deferDelay <= 0 {
			deferDelay = 2500 * time.Millisecond
		}
		if !a.config.AutoDefer {
			deferDelay = 2800 * time.Millisecond
		}

		select {
		case resp := <-gdkCtx.ResponseChannel():
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case <-time.After(deferDelay):
			select {
			case resp := <-gdkCtx.ResponseChannel():
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			default:
				_ = gdkCtx.Defer(false)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(types.InteractionResponse{
					Type: types.CallbackTypeDeferredChannelMessageWithSource,
				})
			}
		}
	})

	a.server = &http.Server{
		Addr:    a.config.WebhookAddr,
		Handler: mux,
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

