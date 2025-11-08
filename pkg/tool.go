package pkg

import "os"

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
