package cleaner

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync/atomic"
	"time"
)

const BACKUP_DIR_NAME = "被删除的文件"

var allDelCount int32
var allFileCount int32
var allDirCount int32
var ignoreCount int32

func DoClean() {
	t := time.Now()
	dirWalk("./")
	logrus.Infof("总扫描文件数:%d", allFileCount)
	logrus.Infof("总扫描文件夹数:%d", allDirCount)
	logrus.Infof("总重复文件被删除数:%d", allDelCount)
	logrus.Infof("总耗时%v", time.Since(t))
}

func dirWalk(dirPath string) {
	if strings.Contains(dirPath, BACKUP_DIR_NAME) {
		return
	}
	log := logrus.WithField("目录", dirPath)
	hidden, err := isFileHidden(dirPath)
	if err != nil {
		log.WithError(err).Info("isFileHidden")
	}
	if hidden {
		return
	}

	fs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		logrus.Panic(err)
	}
	atomic.AddInt32(&allDirCount, 1)

	hash2files := make(map[string][]os.FileInfo, 0)
	for _, file := range fs {
		if file.IsDir() {
			dirWalk(dirPath + file.Name() + "/")
		} else {
			if path.Ext(file.Name()) == ".log" {
				atomic.AddInt32(&ignoreCount, 1)
				continue
			}
			name := dirPath + file.Name()
			md5, err := md5File(name)
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
			backupDir := "./" + BACKUP_DIR_NAME + "/" + dirPath[2:]
			createDirIfNoExist(backupDir)
			err = os.Rename(dirPath+file.Name(), backupDir+file.Name())
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
