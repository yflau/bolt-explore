//+build mage

package main

import (
	"os"
	"fmt"
	"unsafe"
	"time"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/boltdb/bolt"
	"github.com/davecgh/go-spew/spew"
	"github.com/magefile/mage/sh"
)

const (
	defaultPageSize = 4096
)

type pgid uint64

type page struct {
	id       pgid
	flags    uint16
	count    uint16
	overflow uint32
	ptr      uintptr
}

// PageInBufferDemo Open 内打开db文件(文件大小!=0)时读取第一个页的时候id=0, pageSize=0, 因此只是读取固定的4096个字节([0x1000]byte), 
// 但其实一个page结构体仅占8+2+2+4+4=20字节，因此读取20也是可以的
//
// $mage pageInBufferDemo
// (*main.page)(0xc0000d6000)({
//  id: (main.pgid) 0,
//  flags: (uint16) 4,
//  count: (uint16) 0,
//  overflow: (uint32) 0,
//  ptr: (uintptr) 0x2ed0cdaed
// })
func PageInBufferDemo() {
	dbfile, err := os.OpenFile("db", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("%v", err)
		return
	}
	var b [20]byte // NOTE: 源码此处读取4096字节
	if _, err := dbfile.ReadAt(b[:], 0); err == nil {
		id := pgid(0)
		pageSize := 0
		page0 := (*page)(unsafe.Pointer(&b[id*pgid(pageSize)]))
		spew.Dump(page0)
		if page0.flags != 0x04 {
			fmt.Println("the first page type should meta, but got", page0.flags)
		}
	}
}

func openDB(readOnly bool) error {
	opts := &bolt.Options{
		ReadOnly: readOnly,
		Timeout: time.Second,
	}
	db, err := bolt.Open("test.db", 0644, opts) // 只读模式
	if err != nil {
		return errors.Errorf("open test.db error: %v", err)
	}

	fmt.Println("open db successful with %v", opts)
	time.Sleep(3 * time.Second) // 等待只读打开执行
	if err := db.Close(); err != nil {
		return errors.Errorf("close test.db error %v", err)
	}

	return nil
}

// OpenWriter 打开Writer后停顿3秒
func OpenWriter() error {
	return openDB(false)
}

// OpenWriter 打开Reader后停顿3秒
func OpenReader() error {
	return openDB(true)
}

// FlockRace Open不能同时打开读写进程，可以同时打开多个读进程
// 
// $mage flockRace
// Error: open test.db error: timeout
// Error: running "mage openReader" failed with exit code 1
func FlockRace() error {
	cmdW := exec.Command("mage", "openWriter")
	err := cmdW.Start()
	if err != nil {
		return err
	}
	time.Sleep(time.Second)
	err = sh.RunV("mage", "openReader")
	if err != nil {
		return err
	}

	err = cmdW.Wait()
	if err != nil {
		return err
	}

	time.Sleep(4*time.Second)
	return nil
}

