package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const pageSize = 4096

func GenerateKey(userID string) []byte {
	hash := md5.Sum([]byte(userID))
	hexHash := hex.EncodeToString(hash[:])
	return []byte(hexHash[:16])
}

func DecryptDatabase(inputPath, outputPath string, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	buf := make([]byte, pageSize)
	for {
		n, err := inFile.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}

		if n == pageSize {
			decryptECB(block, buf)
		}

		if _, err := outFile.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	}

	return nil
}

func decryptECB(block cipher.Block, data []byte) {
	blockSize := block.BlockSize()
	for i := 0; i < len(data); i += blockSize {
		block.Decrypt(data[i:i+blockSize], data[i:i+blockSize])
	}
}
