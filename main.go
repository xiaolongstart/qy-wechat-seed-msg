package main

import (
	"fmt"
	"io"
	"net/http"
	"syscall"
	"unsafe"

	"github.com/atotto/clipboard"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	findWindow          = user32.NewProc("FindWindowW")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
	showWindow          = user32.NewProc("ShowWindow")
	keybd_event         = user32.NewProc("keybd_event")
)

func main() {
	// cmdPath := `C:\Program Files (x86)\WXWork\WXWork.exe`
	// err := exec.Command(cmdPath).Start()
	// if err != nil {
	// 	fmt.Printf("启动WXWork失败: %v\n", err)
	// 	return
	// }
	// time.Sleep(time.Second * 10)
	// hwnd := robotgo.FindWindow("企业微信")
	// if hwnd == 0 {
	// 	fmt.Println("未找到企业微信窗口")
	// 	return
	// }
	hwnd, err := FindWindow("", "企业微信")
	if err != nil {
		fmt.Println("未找到企业微信窗口")
		panic(err)
	}
	fmt.Println(hwnd)

	http.HandleFunc("/msg", func(w http.ResponseWriter, r *http.Request) {
		SetForegroundWindow(hwnd)
		// 发送一个tab按键
		KeybdEvent(0x09, 0x01)
		// 发送两个退格键
		KeybdEvent(0x08, 0x01)
		KeybdEvent(0x08, 0x01)
		text, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(text))
		// 利用粘贴板将收到的消息粘贴板
		clipboard.WriteAll(string(text))
		// 按下crtl+v
		KeybdEvent(0x11, 0x00)
		KeybdEvent(0x56, 0x01)
		// 释放crtl
		KeybdEvent(0x11, 0x0002)
		// 按下回车
		KeybdEvent(0x0D, 0x01)
	})
	fmt.Println("服务启动 端口8080")
	http.ListenAndServe(":8080", nil)
}

func FindWindow(className, windowName string) (hwnd syscall.Handle, err error) {
	var cname, wname *uint16
	if className != "" {
		cname, err = syscall.UTF16PtrFromString(className)
		if err != nil {
			return 0, err
		}
	}
	if windowName != "" {
		wname, err = syscall.UTF16PtrFromString(windowName)
		if err != nil {
			return 0, err
		}
	}
	r1, _, e1 := syscall.SyscallN(findWindow.Addr(), uintptr(unsafe.Pointer(cname)), uintptr(unsafe.Pointer(wname)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	hwnd = syscall.Handle(r1)
	return
}

func SetForegroundWindow(hwnd syscall.Handle) (err error) {
	r1, _, e1 := syscall.SyscallN(setForegroundWindow.Addr(), uintptr(hwnd), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func ShowWindow(hwnd syscall.Handle, cmd int) (err error) {
	r1, _, e1 := syscall.SyscallN(showWindow.Addr(), uintptr(hwnd), uintptr(cmd), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func KeybdEvent(key byte, down uint8) (err error) {
	r1, _, e1 := syscall.SyscallN(keybd_event.Addr(), uintptr(key), uintptr(0), uintptr(down), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
