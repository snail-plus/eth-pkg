package secure

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func GetMd5String(inputStr string) string {
	h := md5.New()
	h.Write([]byte(inputStr))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func GetFileMd5String(filePath string) (result string) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Open", err)
		return
	}

	defer f.Close()

	md5hash := md5.New()
	if _, err := io.Copy(md5hash, f); err != nil {
		fmt.Println("Copy", err)
		return
	}

	return fmt.Sprintf("%x", md5hash.Sum(nil))
}
