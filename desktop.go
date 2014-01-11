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
