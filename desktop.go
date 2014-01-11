package walk

import (
	"github.com/lxn/win"
)

func GetDesktopSize() Size {
	desktopHWND := win.GetDesktopWindow()
	var rect win.RECT
	win.GetClientRect(desktopHWND, &rect)

	return Size{int(rect.Right), int(rect.Bottom)}
}

func SetWindowToDesktopCenter(w Window) {
	s := GetDesktopSize()
	w.SetX((s.Width - w.Width()) / 2)
	w.SetY((s.Height - w.Height()) / 2)
}
