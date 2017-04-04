package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	logChecksum, err := ioutil.ReadFile("log/checksum")
	if err != nil {
		fmt.Println(err)
	}

	str := string(logChecksum) // convert content to a 'string'

	if b, err := ComputeMd5("log/log"); err != nil {
		fmt.Printf("Err: %v", err)
	} else {
		md5Result := hex.EncodeToString(b)
		if str == md5Result {
			fmt.Println("Log integrity OK.")
		} else {
			fmt.Println("File Tampering detected.")
		}
	}
}

func ComputeMd5(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}
