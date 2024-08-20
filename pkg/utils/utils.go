// Contiene varie utilità
package utils

import (
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Genera una stringa random con una lunghezza specificata nel parametro n
func RandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ@#$%&*1234567890"

	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Converte la prima lettera di una stringa in maiuscolo, rendendola prima tutta in minuscolo
func FirstUpper(s string) string {
	s = strings.ToLower(s)
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Converte una stringa in un int16
func StrToInt16(s string) (int16, error) {
	v, err := strconv.ParseInt(s, 10, 16)
	return int16(v), err
}

// Converte una stringa in minuscolo
func Lower(s string) string {
	return strings.ToLower(s)
}

// LevelStringToSlog converts a log level string to slog.Level(int)
func LevelStringToSlog(level string) slog.Level {
	switch Lower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "fatal":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}

// ClearTerminal pulisce il terminale in base al sistema operativo (windows e linux implementati)
var clear = make(map[string]func()) //create a map for storing clear funcs

func init() {
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func ClearTerminal() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS ritorna linux, windows, etc.
	if ok {                          //Se il sistema operativo è supportato
		value()
	} else { //unsupported platform
		slog.Error("Sistema operativo non supportato. Non posso pulire il terminale.")
	}
}
