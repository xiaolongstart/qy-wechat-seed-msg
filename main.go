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
	bringWindowToTop    = user32.NewProc("BringWindowToTop")
)

const (
	// 用于显示窗口的标志
	SWP_NOMOVE     = 0x0002
	SWP_NOSIZE     = 0x0001
	SWP_SHOWWINDOW = 0x0040
	SWP_RESTORE    = 0x00000009 // 恢复最小化窗口
	HWND_TOPMOST   = -1
	// 恢复窗口的命令
	SW_RESTORE = 9
)

func main() {
	hwnd, err := FindWindow("", "企业微信")
	if err != nil {
		fmt.Println("未找到企业微信窗口")
		panic(err)
	}
	fmt.Println("企业微信窗口句柄:", hwnd)
	http.HandleFunc("/msg", func(w http.ResponseWriter, r *http.Request) {
		// 先恢复最小化窗口
		ShowWindow(hwnd, SW_RESTORE)

		// 聚焦窗口并确保它处于最前面
		SetForegroundWindow(hwnd)
		BringWindowToTop(hwnd)

		// 获取POST请求的消息内容
		text, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "读取请求体失败", http.StatusInternalServerError)
			return
		}
		fmt.Println("接收到的消息:", string(text))

		// 利用粘贴板将收到的消息写入剪贴板
		err = clipboard.WriteAll(string(text))
		if err != nil {
			http.Error(w, "剪贴板操作失败", http.StatusInternalServerError)
			return
		}

		// 模拟 Ctrl+V 粘贴操作
		KeybdEvent(0x11, 0) // 按下 Ctrl
		KeybdEvent(0x56, 0) // 按下 V
		KeybdEvent(0x56, 2) // 松开 V
		KeybdEvent(0x11, 2) // 松开 Ctrl

		// 模拟回车键
		KeybdEvent(0x0D, 0) // 按下回车
		KeybdEvent(0x0D, 2) // 松开回车

		// 响应请求
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("消息已发送到企业微信"))
	})

	fmt.Println("服务启动，监听端口8080...")
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

func BringWindowToTop(hwnd syscall.Handle) (err error) {
	// 使用 BringWindowToTop 强制将窗口置于前台
	r1, _, e1 := syscall.SyscallN(bringWindowToTop.Addr(), uintptr(hwnd), 0, 0)
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
	r1, _, e1 := syscall.SyscallN(keybd_event.Addr(), uintptr(key), uintptr(0), uintptr(down))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
