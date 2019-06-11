// +build windows

package main

import (
	"path/filepath"
	"syscall"
	
	"github.com/lxn/win"
)

func getHostsPath() string {
    buf := make([]uint16, win.MAX_PATH)
	res := win.SHGetSpecialFolderPath(win.HWND(0), &buf[0], win.CSIDL_SYSTEM, false)
	if !res {
		panic("hosts 파일 위치를 불러오지 못했습니다")
	}

	return filepath.Join(syscall.UTF16ToString(buf), "drivers", "etc", "hosts")
}