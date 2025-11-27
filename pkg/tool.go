package pkg

import (
	"bytes"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
	"os"
)

func isPathExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateDirIfNotExist(dirPath string) error {
	exist, err := isPathExist(dirPath)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func DetectAndConvertToUTF8(b []byte) (string, error) {
	if len(b) == 0 {
		return "", nil
	}

	// 自动检测编码
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(b)
	if err != nil {
		return string(b), err
	}

	var enc encoding.Encoding

	switch result.Charset {
	case "UTF-8":
		return string(b), nil

	case "GB-18030", "GBK", "GB2312":
		enc = simplifiedchinese.GBK

	case "UTF-16LE":
		enc = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)

	case "UTF-16BE":
		enc = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)

	case "ISO-8859-1":
		enc = charmap.ISO8859_1

	default:
		// 如果未知，则假设 UTF-8（安全兜底）
		return string(b), nil
	}

	reader := transform.NewReader(bytes.NewReader(b), enc.NewDecoder())
	decoded, err := ioutil.ReadAll(reader)
	if err != nil {
		return string(b), err
	}

	return string(decoded), nil
}
