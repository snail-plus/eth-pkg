package secure

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
)

func GetMd5String(inputStr string) string {
	h := md5.New()
	h.Write([]byte(inputStr))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func GetFileMd5String(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return GetMd5String(string(content)), nil
}
