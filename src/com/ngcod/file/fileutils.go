package file

import (
	"container/list"
	"io"
	"os"
	"path/filepath"
)

var listFiles *list.List
var showSub bool

func ListFiles(src string, isShowSub bool) {
	showSub = isShowSub
	filepath.Walk(src, WalkFunc)
}

func WalkFunc(path string, fileInfo os.FileInfo, err error) error {
	if fileInfo == nil {
		return nil
	}
	if fileInfo.Name() == path {
		return nil
	}

	if fileInfo.IsDir() {
		if showSub {
			ListFiles(fileInfo.Name(), showSub)
		}

	} else {
		listFiles.PushBack(fileInfo)
	}
	return nil
}

func CopyFile(oldpath, newpath string) error {
	file, err := os.OpenFile(oldpath, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer file.Close()

	fileDest, errDest := os.OpenFile(newpath, os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	if errDest != nil {
		return errDest
	}
	defer fileDest.Close()

	var bytes []byte = make([]byte, 1024)
	for {
		size, errRead := file.Read(bytes)
		if errRead == io.EOF {
			if size != 0 {
				fileDest.Write(bytes[:size])
			}
			break
		}
		fileDest.Write(bytes[:size])
	}
	return nil
}
