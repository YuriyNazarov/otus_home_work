//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
}

type getUsersTestData struct {
	input          string
	expectedResult []User
	expectedErr    error
	domain         string
}

func TestGetUsers(t *testing.T) {
	var emptyUsers []User
	dataSet := []getUsersTestData{
		{
			input: `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`,
			expectedResult: []User{
				{Email: "mLynch@broWsecat.com"},
				{Email: "RoseSmith@Browsecat.com"},
				{Email: "nulla@Linktype.com"},
			},
			expectedErr: nil,
			domain:      "com",
		},
		{
			input: `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`,
			expectedResult: emptyUsers,
			expectedErr:    nil,
			domain:         "no_such_domain",
		},
		{
			input: `{{"----""Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`,
			expectedResult: emptyUsers,
			expectedErr:    errInvalidJSON,
			domain:         "gov",
		},
	}

	for i, data := range dataSet {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			resultC, errC := getUsers(bytes.NewBufferString(data.input), data.domain)
			var (
				result []User
				err    error
			)
		Loop:
			for {
				select {
				case user, ok := <-resultC:
					if !ok {
						break Loop
					}
					result = append(result, user)
				case err = <-errC:
					if err != nil {
						break Loop
					}
				}
			}
			require.Equal(t, data.expectedErr, err)
			require.Equal(t, data.expectedResult, result)
		})
	}
}

func BenchmarkGetDomainStat(b *testing.B) {
	r, _ := zip.OpenReader("testdata/users.dat.zip")
	defer r.Close()
	data, _ := r.File[0].Open()
	for i := 0; i < b.N; i++ {
		GetDomainStat(data, "biz")
	}
}

func BenchmarkGetUsers(b *testing.B) {
	r, _ := zip.OpenReader("testdata/users.dat.zip")
	defer r.Close()
	data, _ := r.File[0].Open()
	for i := 0; i < b.N; i++ {
		uC, eC := getUsers(data, "biz")
	L:
		for {
			select {
			case _, ok := <-uC:
				if !ok {
					break L
				}
			case err := <-eC:
				if err != nil {
					break L
				}
			}
		}
	}
}

func BenchmarkCountDomains(b *testing.B) {
	r, _ := zip.OpenReader("testdata/users.dat.zip")
	domain := "biz"
	defer r.Close()
	data, _ := r.File[0].Open()
	usersC, errC := getUsers(data, domain)
	for i := 0; i < b.N; i++ {
		_, err := countDomains(usersC, errC, domain)
		require.NoError(b, err)
	}
}
