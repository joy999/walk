// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package walk

import (
	"unsafe"
)

import (
	"github.com/lxn/win"
)

const mainWindowWindowClass = `\o/ Walk_MainWindow_Class \o/`

func init() {
	MustRegisterWindowClass(mainWindowWindowClass)
}

type MainWindow struct {
	FormBase
	windowPlacement *win.WINDOWPLACEMENT
	menu            *Menu
	toolBar         *ToolBar
	statusBar       *StatusBar
}

func NewMainWindowWithStyle(style, styleEx uint32) (*MainWindow, error) {
	mw := new(MainWindow)

	if style == 0 {
		style = win.WS_OVERLAPPEDWINDOW
	}

	if styleEx == 0 {
		styleEx = win.WS_EX_CONTROLPARENT
	}

	if err := InitWindow(
		mw,
		nil,
		mainWindowWindowClass,
		style,
		styleEx); err != nil {

		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			mw.Dispose()
		}
	}()

	mw.SetPersistent(true)

	var err error

	if mw.menu, err = newMenuBar(); err != nil {
		return nil, err
	}
	if !win.SetMenu(mw.hWnd, mw.menu.hMenu) {
		return nil, lastError("SetMenu")
	}

	if mw.toolBar, err = NewToolBar(mw); err != nil {
		return nil, err
	}
	mw.toolBar.parent = nil
	mw.Children().Remove(mw.toolBar)
	mw.toolBar.parent = mw
	win.SetParent(mw.toolBar.hWnd, mw.hWnd)

	if mw.statusBar, err = NewStatusBar(mw); err != nil {
		return nil, err
	}
	mw.statusBar.parent = nil
	mw.Children().Remove(mw.statusBar)
	mw.statusBar.parent = mw
	win.SetParent(mw.statusBar.hWnd, mw.hWnd)

	// This forces display of focus rectangles, as soon as the user starts to type.
	mw.SendMessage(win.WM_CHANGEUISTATE, win.UIS_INITIALIZE, 0)

	succeeded = true

	return mw, nil
}

func NewMainWindow() (*MainWindow, error) {
	mw := new(MainWindow)

	if err := InitWindow(
		mw,
		nil,
		mainWindowWindowClass,
		win.WS_OVERLAPPEDWINDOW,
		win.WS_EX_CONTROLPARENT); err != nil {

		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			mw.Dispose()
		}
	}()

	mw.SetPersistent(true)

	var err error

	if mw.menu, err = newMenuBar(); err != nil {
		return nil, err
	}
	if !win.SetMenu(mw.hWnd, mw.menu.hMenu) {
		return nil, lastError("SetMenu")
	}

	if mw.toolBar, err = NewToolBar(mw); err != nil {
		return nil, err
	}
	mw.toolBar.parent = nil
	mw.Children().Remove(mw.toolBar)
	mw.toolBar.parent = mw
	win.SetParent(mw.toolBar.hWnd, mw.hWnd)

	if mw.statusBar, err = NewStatusBar(mw); err != nil {
		return nil, err
	}
	mw.statusBar.parent = nil
	mw.Children().Remove(mw.statusBar)
	mw.statusBar.parent = mw
	win.SetParent(mw.statusBar.hWnd, mw.hWnd)

	// This forces display of focus rectangles, as soon as the user starts to type.
	mw.SendMessage(win.WM_CHANGEUISTATE, win.UIS_INITIALIZE, 0)

	succeeded = true

	return mw, nil
}

func (mw *MainWindow) Menu() *Menu {
	return mw.menu
}

func (mw *MainWindow) ToolBar() *ToolBar {
	return mw.toolBar
}

func (mw *MainWindow) StatusBar() *StatusBar {
	return mw.statusBar
}

func (mw *MainWindow) ClientBounds() Rectangle {
	bounds := mw.FormBase.ClientBounds()

	if mw.toolBar.Actions().Len() > 0 {
		tlbBounds := mw.toolBar.Bounds()

		bounds.Y += tlbBounds.Height
		bounds.Height -= tlbBounds.Height
	}

	if mw.statusBar.Visible() {
		bounds.Height -= mw.statusBar.Height()
	}

	return bounds
}

func (mw *MainWindow) SetVisible(visible bool) {
	if visible {
		win.DrawMenuBar(mw.hWnd)

		if mw.clientComposite.layout != nil {
			mw.clientComposite.layout.Update(false)
		}
	}

	mw.FormBase.SetVisible(visible)
}

func (mw *MainWindow) Fullscreen() bool {
	return win.GetWindowLong(mw.hWnd, win.GWL_STYLE)&win.WS_OVERLAPPEDWINDOW == 0
}

func (mw *MainWindow) SetFullscreen(fullscreen bool) error {
	if fullscreen == mw.Fullscreen() {
		return nil
	}

	if fullscreen {
		var mi win.MONITORINFO
		mi.CbSize = uint32(unsafe.Sizeof(mi))

		if mw.windowPlacement == nil {
			mw.windowPlacement = new(win.WINDOWPLACEMENT)
		}

		if !win.GetWindowPlacement(mw.hWnd, mw.windowPlacement) {
			return lastError("GetWindowPlacement")
		}
		if !win.GetMonitorInfo(win.MonitorFromWindow(
			mw.hWnd, win.MONITOR_DEFAULTTOPRIMARY), &mi) {

			return newError("GetMonitorInfo")
		}

		if err := mw.ensureStyleBits(win.WS_OVERLAPPEDWINDOW, false); err != nil {
			return err
		}

		if r := mi.RcMonitor; !win.SetWindowPos(
			mw.hWnd, win.HWND_TOP,
			r.Left, r.Top, r.Right-r.Left, r.Bottom-r.Top,
			win.SWP_FRAMECHANGED|win.SWP_NOOWNERZORDER) {

			return lastError("SetWindowPos")
		}
	} else {
		if err := mw.ensureStyleBits(win.WS_OVERLAPPEDWINDOW, true); err != nil {
			return err
		}

		if !win.SetWindowPlacement(mw.hWnd, mw.windowPlacement) {
			return lastError("SetWindowPlacement")
		}

		if !win.SetWindowPos(mw.hWnd, 0, 0, 0, 0, 0, win.SWP_FRAMECHANGED|win.SWP_NOMOVE|
			win.SWP_NOOWNERZORDER|win.SWP_NOSIZE|win.SWP_NOZORDER) {

			return lastError("SetWindowPos")
		}
	}

	return nil
}

func (mw *MainWindow) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_SIZE, win.WM_SIZING:
		cb := mw.ClientBounds()

		mw.toolBar.SetBounds(Rectangle{0, 0, cb.Width, mw.toolBar.Height()})
		mw.statusBar.SetBounds(Rectangle{0, cb.Height, cb.Width, mw.statusBar.Height()})
	}

	return mw.FormBase.WndProc(hwnd, msg, wParam, lParam)
}
