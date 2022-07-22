package hw10programoptimization

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/valyala/fastjson"
)

type User struct {
	Email string
}

type DomainStat map[string]int

var errInvalidJSON = errors.New("error occupied on parsing json")

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	userC, errC := getUsers(r, domain)
	return countDomains(userC, errC, domain)
}

func getUsers(r io.Reader, domain string) (<-chan User, <-chan error) {
	linesC := readLine(r, domain)
	usersChan := make(chan User, 10)
	errChan := make(chan error)

	go func() {
		var (
			user User
			p    fastjson.Parser
			val  *fastjson.Value
			err  error
		)
	LinesRead:
		for {
			line, ok := <-linesC
			if !ok {
				break LinesRead
			}
			val, err = p.Parse(line)
			if err != nil {
				errChan <- errInvalidJSON
				break LinesRead
			}
			user.Email = string(val.GetStringBytes("Email"))
			usersChan <- user
		}
		close(usersChan)
		close(errChan)
	}()
	return usersChan, errChan
}

func countDomains(usersC <-chan User, errC <-chan error, domain string) (DomainStat, error) {
	var (
		err  error
		ok   bool
		user User
	)
	result := make(DomainStat, 100)

UserRead:
	for {
		select {
		case user, ok = <-usersC:
			if !ok {
				break UserRead
			}
			if strings.HasSuffix(user.Email, domain) {
				result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]++
			}

		case err = <-errC:
			if err != nil {
				return DomainStat{}, err
			}
		}
	}
	return result, nil
}

func readLine(r io.Reader, domain string) <-chan string {
	outChan := make(chan string, 10)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	go func() {
		for scanner.Scan() {
			if !bytes.Contains(scanner.Bytes(), []byte(domain)) {
				continue
			}
			outChan <- string(scanner.Bytes())
		}
		close(outChan)
	}()
	return outChan
}
