package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"log"
)

type Logger struct {
	Filename   string
	Prefix     string
	MaxBytes   int64
	MaxBackups int

	size int64
	file *os.File
	mu   sync.Mutex
}

var lineBeginRegex = regexp.MustCompile(`(?m)^`)
var backupSeqRegex = regexp.MustCompile(`\.(\w+)$`)

func (l *Logger) Write(p []byte) (int, error) {
	if l.Prefix != "" {
		p = lineBeginRegex.ReplaceAll(p, []byte(l.Prefix))
	}

	if l.Filename == "" {
		log.Print(p)
	} else {
		l.mu.Lock()
		defer l.mu.Unlock()

		f, err := l.getCurrentFile(len(p))
		if err != nil {
			log.Println(err)
			return 0, err
		}

		n, err := f.Write(p)
		l.size += int64(n)
	}

	return len(p), nil
}

func (l *Logger) getCurrentFile(bytes int) (*os.File, error) {
	if l.file != nil {
		if l.size+int64(bytes) < l.MaxBytes {
			return l.file, nil
		}

		l.file.Close()
		l.file = nil
		l.size = 0

		l.rotateBackups()
	}

	return l.openExistOrNewFile()
}

func (l *Logger) openExistOrNewFile() (*os.File, error) {
	os.MkdirAll(path.Dir(l.Filename), 0777)

	file, err := os.OpenFile(l.Filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	l.file = file
	l.size = stat.Size()

	return l.file, nil
}

func (l *Logger) rotateBackups() {
	backups, _ := filepath.Glob(l.Filename + ".*")
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	for _, backup := range backups {
		m := backupSeqRegex.FindStringSubmatch(backup)
		if len(m) == 0 {
			log.Printf("Invalid filename:%s, ignore", backup)
			continue
		}
		backupSeqID, err := strconv.Atoi(m[1])
		if err != nil {
			log.Printf("Invalid filename:%s, ignore", backup)
			continue
		}
		backupSeqID = backupSeqID + 1
		if backupSeqID > l.MaxBackups {
			os.Remove(backup)
			continue
		}
		if err := os.Rename(backup, fmt.Sprintf("%s.%d", l.Filename, backupSeqID)); err != nil {
			log.Println(err)
		}
	}

	if l.MaxBackups > 0 {
		if err := os.Rename(l.Filename, l.Filename+".1"); err != nil {
			log.Println(err)
		}
	} else {
		if err := os.Remove(l.Filename); err != nil {
			log.Println(err)
		}
	}
}
