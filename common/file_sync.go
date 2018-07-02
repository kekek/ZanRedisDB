package common

import (
	//"github.com/Redundancy/go-sync"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
)

var runningCh chan struct{}

func init() {
	runningCh = make(chan struct{}, 2)
}

var rsyncLimit = int64(51200)

func SetRsyncLimit(limit int64) {
	atomic.StoreInt64(&rsyncLimit, limit)
}

func RunFileSync(remote string, srcPath string, dstPath string, stopCh chan struct{}) error {
	select {
	case runningCh <- struct{}{}:
	case <-stopCh:
		return ErrStopped
	}
	defer func() {
		select {
		case <-runningCh:
		default:
		}
	}()

	var cmd *exec.Cmd
	if filepath.Base(srcPath) == filepath.Base(dstPath) {
		dir := filepath.Dir(dstPath)
		os.MkdirAll(dir, DIR_PERM)
	} else {
		os.MkdirAll(dstPath, DIR_PERM)
	}
	if remote == "" {
		log.Printf("copy local :%v to %v\n", srcPath, dstPath)
		cmd = exec.Command("cp", "-rp", srcPath, dstPath)
	} else {
		log.Printf("copy from remote :%v/%v to local: %v\n", remote, srcPath, dstPath)
		// limit rate in kilobytes
		limitStr := fmt.Sprintf("--bwlimit=%v", atomic.LoadInt64(&rsyncLimit))
		cmd = exec.Command("rsync", "-avP", limitStr,
			"rsync://"+remote+"/"+srcPath, dstPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		log.Printf("cmd %v error: %v", cmd, err)
	}
	return err
}
