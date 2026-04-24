package util

import (
	"golang.org/x/crypto/blowfish"

	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
)

/* Calculate MD5 */
func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)

	return hex.EncodeToString(hash[:])
}

/* Calculate CRC32 checksum */
func CalculateCRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

/* Encrypt data using Blowfish ECB mode */
func EncryptBlowfishECB(key []byte, data []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	dataLen := len(data)

	if dataLen%8 != 0 {
		return nil, fmt.Errorf("data length must be a multiple of 8")
	}

	cData := make([]byte, dataLen)
	for i := 0; i < dataLen; i += 8 {
		block.Encrypt(cData[i:i+8], data[i:i+8])
	}

	return cData, err
}

/* Decrypt data using Blowfish ECB mode */
func DecryptBlowfishECB(key []byte, cData []byte) ([]byte, error) {
	block, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cDataLen := len(cData)
	if cDataLen%8 != 0 {
		return nil, fmt.Errorf("cipher data length must be a multiple of 8")
	}

	data := make([]byte, cDataLen)
	for i := 0; i < cDataLen; i += 8 {
		block.Decrypt(data[i:i+8], cData[i:i+8])
	}

	return data, err
}
