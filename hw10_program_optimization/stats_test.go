//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}`
	//todo temp
	//t.Run("find 'co123m'", func(t *testing.T) {
	//	result, err := readLines(bytes.NewBufferString(data))
	//	fmt.Println(err)
	//	for _, s := range result {
	//		fmt.Println("str is:", s)
	//	}
	//})

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
		getUsers(data)
	}
}

func BenchmarkCountDomains(b *testing.B) {
	r, _ := zip.OpenReader("testdata/users.dat.zip")
	defer r.Close()
	data, _ := r.File[0].Open()
	users, cnt, _ := getUsers(data)
	for i := 0; i < b.N; i++ {
		countDomains(users, cnt, "biz")
	}
}
