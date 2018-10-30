package handler

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func GetENV(env string) string {
	env_value := os.Getenv(env)
	if env_value == "" {
		fmt.Println("FATAL: NEED ENV", env)
		fmt.Println("Exit...........")
		os.Exit(2)
	}
	fmt.Println("ENV:", env, env_value)
	return env_value
}

func CheckFileIsExist(filename string) (exist bool, err error) {
	exist = true
	if _, err = os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist, err
}

func getmd5string(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func Getguid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getmd5string(base64.URLEncoding.EncodeToString(b))
}

func CheckDBname(dbname string) string {
	lenth := len(dbname)
	if lenth == 0 {
		return "d" + Getguid()[0:15]
	}
	if lenth > 32 {
		dbname = dbname[:32]
	}

	//不合法字符统一替换为"_"
	for i, v := range dbname {
		// A:65 Z:90 a:97 z:122 0:48 9:57 _:95 $:36
		if (v > 64 && v < 91) || (v > 96 && v < 123) || (v > 47 && v < 58) || v == 95 || v == 36 {
			if i == 0 && (v > 47 && v < 58) {
				//not allow dbname begins with number
				dbname = strings.Replace(dbname, string(dbname[i]), "_", 1)
			}
		} else {
			dbname = strings.Replace(dbname, string(dbname[i]), "_", 1)
		}

	}

	return dbname
}
