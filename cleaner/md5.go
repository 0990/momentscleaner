package cleaner

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"sort"
	"sync"
)

func MD5Bytes(s []byte) string {
	ret := md5.Sum(s)
	return hex.EncodeToString(ret[:])
}

//计算字符串MD5值
func MD5(s string) string {
	return MD5Bytes([]byte(s))
}

//计算文件MD5值
func MD5File(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return MD5Bytes(data), nil
}

type md5result struct {
	file string
	md5  string
	err  error
}

//多个文件计算时，result = md5(md5(file1)+md5(file2))
func MD5Files(files ...string) (string, error) {
	length := len(files)

	if length == 0 {
		return "", errors.New("input param error")
	}

	if length == 1 {
		return MD5File(files[0])
	}

	var wg sync.WaitGroup
	var m sync.Map

	wg.Add(length)
	for _, v := range files {
		file := v
		go func() {
			md5, err := MD5File(file)
			md5result := &md5result{
				file: file,
				md5:  md5,
				err:  err,
			}
			m.Store(file, md5result)
			wg.Done()
		}()
	}

	var sortfiles []string
	sortfiles = append(sortfiles, files...)
	sort.Strings(sortfiles)
	wg.Wait()

	var sum string
	for _, file := range sortfiles {
		if v, ok := m.Load(file); ok {
			result := v.(*md5result)
			if result.err != nil {
				return "", result.err
			}
			sum += result.md5
		}
	}

	return MD5(sum), nil
}
