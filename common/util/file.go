package util

import (
	"os"
)

func CheckFilePath(filePath string) error {
	var err error
	_, err = os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
		}
		return err
	}

	return nil
}
