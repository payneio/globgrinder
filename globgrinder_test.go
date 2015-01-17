package globgrinder

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestIntegration(t *testing.T) {

	defer clean()
	err := prepare()
	require.Nil(t, err)

	pattern := "./tmp.in/*.txt"
	outDir := "./tmp.out"

	gg, err := New(pattern, outDir)
	require.Nil(t, err)

	process := make(chan string)
	done := make(chan bool)

	go gg.Run(process, done)

	assert.Equal(t, <-process, "tmp.in/a.txt.grinding")
	done <- true
	assert.Equal(t, <-process, "tmp.in/b.txt.grinding")
}

func TestUniqueIncrementedPath(t *testing.T) {

	var err error

	defer clean()
	err = prepare()
	require.Nil(t, err)

	pattern := "./tmp.in/*.txt"
	outDir := "./tmp.out"

	_, err = New(pattern, outDir)
	require.Nil(t, err)

	err = ioutil.WriteFile("./tmp.out/c.txt", []byte("test"), 0666)
	require.Nil(t, err)
	var s string
	s = uniqueIncrementedPath("tmp.out/c.txt")
	assert.Equal(t, "tmp.out/c.txt.1", s)

	err = ioutil.WriteFile("./tmp.out/c.txt.8", []byte("test"), 0666)
	require.Nil(t, err)
	s = uniqueIncrementedPath("tmp.out/c.txt")
	assert.Equal(t, "tmp.out/c.txt.9", s)

	err = ioutil.WriteFile("./tmp.out/c.txt.320", []byte("test"), 0666)
	require.Nil(t, err)
	s = uniqueIncrementedPath("tmp.out/c.txt")
	assert.Equal(t, "tmp.out/c.txt.321", s)

}

func prepare() error {
	if err := os.MkdirAll("./tmp.in", 0777); err != nil {
		return err
	}
	if err := ioutil.WriteFile("./tmp.in/a.txt", []byte("hello"), 0666); err != nil {
		return err
	}
	if err := ioutil.WriteFile("./tmp.in/b.txt", []byte("world"), 0666); err != nil {
		return err
	}
	if err := os.Mkdir("./tmp.out", 0777); err != nil {
		return err
	}
	return nil
}

func createTestFile(name string) error {
	err := ioutil.WriteFile(name, []byte("test"), 0666)
	return err
}

func clean() {
	os.RemoveAll("./tmp.in")
	os.RemoveAll("./tmp.out")
}
