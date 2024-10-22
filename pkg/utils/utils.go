// Contents: utility functions for the project
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

// RandomString generates a random string of length n made of letters (lower and uppercase) and numbers
func RandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// FirstUpper converts the first letter of a string to uppercase after converting the whole string to lowercase
func FirstUpper(s string) string {
	s = strings.ToLower(s)
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// StrToInt16 converts a string to int16
func StrToInt16(s string) (int16, error) {
	v, err := strconv.ParseInt(s, 10, 16)
	return int16(v), err
}

// Lower converts a string to lowercase
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

// ClearTerminal clears the terminal screen based on the O.S. (linux and windows implemented)
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
	value, ok := clear[runtime.GOOS] //runtime.GOOS returns linux, windows, etc.
	if ok {                          //If the O.S. is supported
		value()
	} else { //unsupported platform
		slog.Error("Sistema operativo non supportato. Non posso pulire il terminale.")
	}
}
