package lib

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"os"
)

func TestResolveOneRedirect(t *testing.T) {
	res := Resolve("http://bit.ly/2hcXx5Z", "http://bit.ly/", 10, false, "Googlebot")
	assert.Equal(t, "https://maddevs.io/", res)
}

func TestResolveTwoRedirects(t *testing.T) {
	res := Resolve("http://ow.ly/Pkxu306YmRs", "http://ow.ly/", 10, false, "Googlebot")
	assert.Equal(t, "https://maddevs.io/", res)
}

func TestResolveSSL(t *testing.T) {
	res := Resolve("https://bit.ly/2hcXx5Z", "http://bit.ly/", 10, false, "Googlebot")
	assert.Equal(t, "https://maddevs.io/", res)

	res = Resolve("https://bit.ly/ID7AM5", "http://bit.ly/", 10, false, "Googlebot")
	assert.Equal(t, "https://www.youtube.com/", res)
}

func TestBackup(t *testing.T) {
	os.Remove("/tmp/res.db")
	err := BackupDatabase(DBFilepath)
	assert.NoError(t, err)
	_, err = os.Stat("/tmp/res.db")
	assert.Equal(t, false, os.IsNotExist(err))
}