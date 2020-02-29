package cleaner

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

func createDirIfNoExist(path string) {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(path, os.ModePerm) //  Everyone can read write and execute
			return
		}
		return
	}
}

func isFileHidden(path string) (bool, error) {

	name := utf16.Encode([]rune(path + "\x00"))

	attributes, err := syscall.GetFileAttributes((*uint16)(unsafe.Pointer(&name[0])))

	if err != nil {

		return false, err

	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil

}

func md5File(filename string) (string, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return "", err
	}
	md5h := md5.New()
	io.Copy(md5h, file)
	return fmt.Sprintf("%x", md5h.Sum([]byte(""))), nil
}
