package file

import (
	"io/ioutil"
	"os"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ReadString(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	return string(content), err
}
