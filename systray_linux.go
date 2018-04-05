// +build linux

package systray

/*
#cgo linux pkg-config: gtk+-3.0 appindicator3-0.1

#include "systray_linux.h"
*/
import "C"
import (
	"crypto/md5"
	"encoding/hex"
	"path/filepath"
	"os"
	"io/ioutil"
)

func nativeLoop() (err error) {
	_, err = C.nativeLoop()
	return
}

func quit() {
	C.quit()
}

// SetIcon sets the systray icon.
// iconBytes should be the content of .ico for windows and .ico/.jpg/.png
// for other platforms.
func SetIcon(iconBytes []byte) (err error) {
	bh := md5.Sum(iconBytes)
	dataHash := hex.EncodeToString(bh[:])
	iconFilePath := filepath.Join(os.TempDir(), "systray_temp_icon_"+dataHash)

	if _, err := os.Stat(iconFilePath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(iconFilePath, iconBytes, 0644); err != nil {
			return err
		}
	}

	_, err = C.setIcon(C.CString(iconFilePath))
	return
}

// SetTitle sets the systray title, only available on Mac.
func SetTitle(title string) (err error) {
	_, err = C.setTitle(C.CString(title))
	return
}

// SetTooltip sets the systray tooltip to display on mouse hover of the tray icon,
// only available on Mac and Windows.
func SetTooltip(tooltip string) (err error) {
	_, err = C.setTooltip(C.CString(tooltip))
	return
}

func addOrUpdateMenuItem(item *MenuItem) (err error) {
	var disabled C.short
	if item.disabled {
		disabled = 1
	}
	var checked C.short
	if item.checked {
		checked = 1
	}
	_, err = C.add_or_update_menu_item(
		C.int(item.id),
		C.CString(item.title),
		C.CString(item.tooltip),
		disabled,
		checked,
	)
	return
}

func addSeparator(id int32) (err error) {
	_, err = C.add_separator(C.int(id))
	return
}

func hideMenuItem(item *MenuItem) (err error) {
	_, err = C.hide_menu_item(
		C.int(item.id),
	)
	return
}

func showMenuItem(item *MenuItem) (err error) {
	_, err = C.show_menu_item(
		C.int(item.id),
	)
	return
}

//export systray_ready
func systray_ready() {
	systrayReady()
}

//export systray_on_exit
func systray_on_exit() {
	systrayExit()
}

//export systray_menu_item_selected
func systray_menu_item_selected(cID C.int) {
	systrayMenuItemSelected(int32(cID))
}