package cleaner

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

const BACKUP_DIR_NAME = "被删除的文件"

var allDelCount int32
var allFileCount int32
var allDirCount int32

func DoClean() {
	t := time.Now()
	dirWalk("./")
	logrus.Infof("总扫描文件数:%d", allFileCount)
	logrus.Infof("总扫描文件夹数:%d", allDirCount)
	logrus.Infof("总重复文件被删除数:%d", allDelCount)
	logrus.Infof("总耗时%v", time.Since(t))
}

func dirWalk(path string) {
	if strings.Contains(path, BACKUP_DIR_NAME) {
		return
	}
	log := logrus.WithField("目录", path)
	hidden, err := isFileHidden(path)
	if err != nil {
		log.WithError(err).Info("isFileHidden")
	}
	if hidden {
		return
	}

	fs, err := ioutil.ReadDir(path)
	if err != nil {
		logrus.Panic(err)
	}
	atomic.AddInt32(&allDirCount, 1)

	hash2files := make(map[string][]os.FileInfo, 0)
	for _, file := range fs {
		if file.IsDir() {
			dirWalk(path + file.Name() + "/")
		} else {
			name := path + file.Name()
			md5, err := MD5File(name)
			if err != nil {
				logrus.Panic(err)
			}
			hash2files[md5] = append(hash2files[md5], file)

			atomic.AddInt32(&allFileCount, 1)
		}
	}

	log.Info("扫描开始")
	var delCount int32
	for _, files := range hash2files {
		if len(files) < 2 {
			continue
		}
		//保留名称最短的文件，其它重复文件删除
		min := len(files[0].Name())
		for i := 1; i < len(files); i++ {
			lname := len(files[i].Name())
			if lname < min {
				min = lname
			}
		}
		for _, file := range files {
			if len(file.Name()) == min {
				log.WithField("filename", file.Name()).Info("保留")
				continue
			}
			backupDir := "./" + BACKUP_DIR_NAME + "/" + path[2:]
			createDirIfNoExist(backupDir)
			err = os.Rename(path+file.Name(), backupDir+file.Name())
			if err != nil {
				logrus.Panic(err)
			}
			delCount++
			atomic.AddInt32(&allDelCount, 1)
			log.WithField("filename", file.Name()).Info("删除")
		}
	}

	if delCount > 0 {
		log.Infof("%d个文件被移除", delCount)
	}
}

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
