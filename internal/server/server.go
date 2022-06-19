package server

import (
	"DB_coursework/internal/utils"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type ServerGen interface {
	Shutdown(context.Context) error
	LogError(string, error)
	LogInfo(string)
	ListenAndServe() error
	CloseLog()
	GetIPAndPort() string
	QueryPass(string, string, *string, *string) error
	GetUuid(string) (map[string]string, error)
	GetCookieTTL(string) (time.Duration, error)
	SetCookie(string, string, string)
}

func LoginCookie(server ServerGen, r *http.Request) (int, string, string, *http.Cookie) {
	var cookie, err = r.Cookie("library-cookie")
	if err != nil {
		return 401, "", "", nil
	}
	cookieVal := cookie.Value
	keys, err := server.GetUuid(cookieVal)
	if err != nil {
		server.LogError("Не удалось получить uuid по cookie", err)
		return 500, "", "", nil
	}
	if len(keys) == 0 {
		return 401, "", "", nil
	}
	ttl, err := server.GetCookieTTL(cookieVal)
	if err != nil {
		return 500, keys["uuid"], keys["isAdmin"], nil
	}
	if ttl < time.Second*1800 {
		newCookie := utils.RandVerificationCode(10)
		server.SetCookie(newCookie, keys["uuid"], keys["isAdmin"])
		return 200, keys["uuid"], keys["isAdmin"], &http.Cookie{
			Name:     "library-cookie",
			Value:    newCookie,
			MaxAge:   7200,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		}
	}
	return 200, keys["uuid"], keys["isAdmin"], nil
}

func LoginPass(server ServerGen, r *http.Request) (string, int, string) {
	url := r.FormValue("url")
	isAdmin := "false"
	table := "end_user"
	if strings.Contains(url, "admin") {
		isAdmin = "true"
		table = "administrator"
	}
	if len(r.Header["Authorization"]) < 1 {
		return "", 400, ""
	}
	b64 := strings.Split(r.Header["Authorization"][0], " ")
	if len(b64) < 2 {
		return "", 400, ""
	}
	decodedb64, err := base64.StdEncoding.DecodeString(b64[1])
	if err != nil {
		return "", 400, ""
	}
	logPass := strings.Split(string(decodedb64), ":")
	login := logPass[0]
	password := logPass[1]
	h2 := sha256.New()
	h2.Write([]byte(password))
	passHash := hex.EncodeToString(h2.Sum(nil))
	var passFromBase string
	var uuid string
	err = server.QueryPass(table, login, &uuid, &passFromBase)
	if err != nil {
		server.LogError("Ошибка при получении пароля и guid из базы", err)
		return "", 500, ""
	}
	if passHash != passFromBase {
		return "", 401, ""
	}
	return isAdmin, 200, uuid
}

func Serve(server ServerGen) {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*20, "the duration for which the server gracefully wait for existing connections to finish")
	flag.Parse()

	shutdownCh := make(chan struct{}, 1)
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGQUIT)
		<-sigint

		graceCtx, graceCancel := context.WithTimeout(context.Background(), wait)
		defer graceCancel()

		if err := server.Shutdown(graceCtx); err != nil {
			server.LogError("can't shutdown", err)
		} else {
			server.LogInfo("server shut down")
		}
		shutdownCh <- struct{}{}
	}()

	server.LogInfo("listening on " + server.GetIPAndPort())
	var err error
	if err = server.ListenAndServe(); err != nil {
		shutdownCh <- struct{}{}
	}

	<-shutdownCh

	if err != http.ErrServerClosed {
		server.LogError("problem while closing server", err)
	}
	server.LogInfo("HTTP server Serve closed")
	server.CloseLog()
}
