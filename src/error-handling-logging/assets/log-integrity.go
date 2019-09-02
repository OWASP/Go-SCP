package main

import (
	"crypto/sha256"
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

	if b, err := ComputeSHA256("log/log"); err != nil {
		fmt.Printf("Err: %v", err)
	} else {
		hash := hex.EncodeToString(b)
		if str == hash {
			fmt.Println("Log integrity OK.")
		} else {
			fmt.Println("File Tampering detected.")
		}
	}
}

func ComputeSHA256(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}
