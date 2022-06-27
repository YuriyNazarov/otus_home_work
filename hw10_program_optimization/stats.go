package hw10programoptimization

import (
	"fmt"
	"github.com/valyala/fastjson"
	"io"
	"strings"
)

type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, count, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, count, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (result users, usersCount int, err error) {
	//content, err := ioutil.ReadAll(r)
	content, err := read(r)
	if err != nil {
		return
	}
	var user User

	lines := strings.Split(content, "\n")
	var (
		p   fastjson.Parser
		val *fastjson.Value
	)
	for i, line := range lines {
		val, err = p.Parse(line)
		if err != nil {
			return result, usersCount, err // так и не понял почему тут ругается компилятор если использовать возврат по имени. ./stats.go:46:4: err is shadowed during return
		}
		user.Email = string(val.GetStringBytes("Email"))
		result[i] = user
		usersCount++
	}
	return
}

func countDomains(u users, usersCount int, domain string) (DomainStat, error) {
	result := make(DomainStat, usersCount/16) // Не уверен является ли это читерством.
	// Рассчет на то что не будет слишком много уникальных доменов подходящих под поиск.
	// При превышении конечно же будут дополнительные аллокации памяти под мапу
	var num int
	for i := 0; i < usersCount; i++ {
		if strings.HasSuffix(u[i].Email, domain) {
			num = result[strings.ToLower(strings.SplitN(u[i].Email, "@", 2)[1])]
			num++
			result[strings.ToLower(strings.SplitN(u[i].Email, "@", 2)[1])] = num
		}
	}
	return result, nil
}

func read(r io.Reader) (string, error) {
	var (
		err     error
		read    int
		bufSize = 1 << 15
	)
	buf := make([]byte, bufSize) // буфер на 32кб
	strBuf := strings.Builder{}

	for {
		read, err = r.Read(buf)
		if err != nil {
			if err == io.EOF {
				strBuf.Write(buf[:read])
				return strBuf.String(), nil
			}
			return "", err
		}
		strBuf.Write(buf[:read])
		buf = make([]byte, bufSize)
	}
}
