//+build mage

package main

import (
	"os"
	"fmt"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
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

// $mage pageInBufferDemo
// (*main.page)(0xc0000d6000)({
//  id: (main.pgid) 0,
//  flags: (uint16) 4,
//  count: (uint16) 0,
//  overflow: (uint32) 0,
//  ptr: (uintptr) 0x2ed0cdaed
// })
// Open 内打开db文件读取第一个页的时候id=0, pageSize=0, 因此只是读取固定的4096个字节, 即[0x1000]byte
// 但其实一个page结构体仅占8+2+2+4+4=20字节，因此读取20也是可以的
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

