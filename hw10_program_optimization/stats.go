package hw10programoptimization

import (
	"github.com/valyala/fastjson"
	"io"
	"strings"
)

type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	userC, errC := getUsers(r)
	return countDomains(userC, errC, domain)
}

func getUsers(r io.Reader) (<-chan User, <-chan error) {
	linesC, errs := readLine(r)
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
			select {
			case line, ok := <-linesC:
				if !ok {
					close(errChan)
					close(usersChan)
					break LinesRead
				}
				val, err = p.Parse(line)
				if err != nil {
					errChan <- err
					close(errChan)
					close(usersChan)
					break LinesRead
				}
				user.Email = string(val.GetStringBytes("Email"))
				usersChan <- user
			case err = <-errs:
				if err != nil {
					errChan <- err
					close(errChan)
					close(usersChan)
				}
			}
		}
	}()
	return usersChan, errChan
}

func countDomains(usersC <-chan User, errC <-chan error, domain string) (DomainStat, error) {
	var (
		num  int
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
				num = result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]
				num++
				result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])] = num
			}
		case err = <-errC:
			if err != nil {
				return DomainStat{}, err
			}
		}
	}
	return result, nil
}

func readLine(r io.Reader) (<-chan string, <-chan error) {
	var (
		err     error
		read    int
		bufSize = 1 << 15
	)
	buf := make([]byte, bufSize) // буфер на 32кб
	strBuf := strings.Builder{}
	outChan := make(chan string, 10)
	errChan := make(chan error)

	go func() {
		for {
			read, err = r.Read(buf)
			if err != nil {
				if err == io.EOF {
					for i := 0; i < read; i++ {
						if buf[i] == '\n' {
							outChan <- strBuf.String()
							strBuf.Reset()
							continue
						}
						strBuf.WriteByte(buf[i])
					}

					outChan <- strBuf.String()
					close(outChan)
					close(errChan)
					break
				} else {
					errChan <- err
					close(outChan)
					close(errChan)
					break
				}

			}
			for i := 0; i < read; i++ {
				if buf[i] == '\n' {
					outChan <- strBuf.String()
					strBuf.Reset()
					continue
				}
				strBuf.WriteByte(buf[i])
			}
			buf = make([]byte, bufSize)
		}
	}()
	return outChan, errChan
}
