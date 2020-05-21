package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var lock sync.Mutex

func FileSavef(format string, arg ...interface{}) error {
	return FileSave(fmt.Sprintf(format, arg))
}

func FileSave(text string) error {
	now := time.Now()
	lock.Lock()
	defer lock.Unlock()
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		_ = os.MkdirAll("./logs", 0776)
	}
	f, err := os.OpenFile(fmt.Sprintf("./logs/%s.log", now.Format("2006-01-02")), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err := f.WriteString(fmt.Sprintf("%s: %s\n", now.Format("2006-01-02 15:04:05"), strings.Trim(text, " \n\r\t"))); err != nil {
		return err
	}
	return nil
}
