package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/jackc/pgx/v4/pgxpool"
	"math/rand"
	"os/exec"
	"strconv"
	"time"
)

//https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func RandVerificationCode(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

func CheckExistence(db *pgxpool.Pool, table, row, value string) (bool, error) {
	var exists bool
	err := db.QueryRow(context.Background(), "select exists (select 1 from "+table+" where "+row+"=$1)", value).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetPassHash(password string) string {
	h2 := sha256.New()
	h2.Write([]byte(password))
	return hex.EncodeToString(h2.Sum(nil))
}

func ExecCmdLine(command string, outb, errb *bytes.Buffer) error {
	outb.Reset()
	errb.Reset()
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stdout = outb
	cmd.Stderr = errb
	return cmd.Run()
}

func GetGenre(genreCode string) string {
	switch genreCode {
	case "0":
		return "sci-fi"
	case "1":
		return "fantasy"
	case "2":
		return "comics"
	case "3":
		return "satire"
	case "4":
		return "crime"
	case "5":
		return "adventure"
	case "6":
		return "historical"
	case "7":
		return "religious"
	case "8":
		return "horror"
	default:
		return "nonfiction"
	}
}

func GetGenreCode(genreCode string) string {
	switch genreCode {
	case "sci-fi":
		return "0"
	case "fantasy":
		return "1"
	case "comics":
		return "2"
	case "satire":
		return "3"
	case "crime":
		return "4"
	case "adventure":
		return "5"
	case "historical":
		return "6"
	case "religious":
		return "7"
	case "horror":
		return "8"
	default:
		return "9"
	}
}

func GetYear(year string) (int, bool) {
	if year == "" {
		return 0, false
	}
	res, err := strconv.Atoi(year)
	if err != nil {
		return 0, false
	}
	return res, true
}
