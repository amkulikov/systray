/*
Package systray is a cross platfrom Go library to place an icon and menu in the
notification area.
Supports Windows, Mac OSX and Linux currently.
Methods can be called from any goroutine except Run(), which should be called
at the very beginning of main() to lock at main thread.
*/
package systray

import (
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	hasStarted = int32(0)
	hasQuit    = int32(0)
)

// MenuItem is used to keep track each menu item of systray
// Don't create it directly, use the one systray.AddMenuItem() returned
type MenuItem struct {
	// ClickedCh is the channel which will be notified when the menu item is clicked
	ClickedCh chan struct{}

	// id uniquely identify a menu item, not supposed to be modified
	id int32
	// title is the text shown on menu item
	title string
	// tooltip is the text shown when pointing to menu item
	tooltip string
	// disabled menu item is grayed out and has no effect when clicked
	disabled bool
	// checked menu item has a tick before the title
	checked bool
}

var (
	systrayReady  func()
	systrayExit   func()
	menuItems     = make(map[int32]*MenuItem)
	menuItemsLock sync.RWMutex

	currentID = int32(-1)
)

// Run initializes GUI and starts the event loop, then invokes the onReady
// callback.
// It blocks until systray.Quit() is called.
// Should be called at the very beginning of main() to lock at main thread.
func Run(onReady func(), onExit func()) error {
	runtime.LockOSThread()
	atomic.StoreInt32(&hasStarted, 1)

	if onReady == nil {
		systrayReady = func() {}
	} else {
		// Run onReady on separate goroutine to avoid blocking event loop
		readyCh := make(chan interface{})
		go func() {
			<-readyCh
			onReady()
		}()
		systrayReady = func() {
			close(readyCh)
		}
	}

	// unlike onReady, onExit runs in the event loop to make sure it has time to
	// finish before the process terminates
	if onExit == nil {
		onExit = func() {}
	}
	systrayExit = onExit

	return nativeLoop()
}

// Quit the systray
func Quit() {
	if atomic.LoadInt32(&hasStarted) == 1 && atomic.CompareAndSwapInt32(&hasQuit, 0, 1) {
		quit()
	}
}

// AddMenuItem adds menu item with designated title and tooltip, returning a channel
// that notifies whenever that menu item is clicked.
//
// It can be safely invoked from different goroutines.
func AddMenuItem(title string, tooltip string) (*MenuItem, error) {
	id := atomic.AddInt32(&currentID, 1)
	item := &MenuItem{nil, id, title, tooltip, false, false}
	item.ClickedCh = make(chan struct{})
	return item, item.update()
}

// AddSeparator adds a separator bar to the menu
func AddSeparator() error {
	return addSeparator(atomic.AddInt32(&currentID, 1))
}

// SetTitle set the text to display on a menu item
func (item *MenuItem) SetTitle(title string) error {
	item.title = title
	return item.update()
}

// SetTooltip set the tooltip to show when mouse hover
func (item *MenuItem) SetTooltip(tooltip string) error {
	item.tooltip = tooltip
	return item.update()
}

// Disabled checkes if the menu item is disabled
func (item *MenuItem) Disabled() bool {
	return item.disabled
}

// Enable a menu item regardless if it's previously enabled or not
func (item *MenuItem) Enable() error {
	item.disabled = false
	return item.update()
}

// Disable a menu item regardless if it's previously disabled or not
func (item *MenuItem) Disable() error {
	item.disabled = true
	return item.update()
}

// Hide hides a menu item
func (item *MenuItem) Hide() error {
	return hideMenuItem(item)
}

// Show shows a previously hidden menu item
func (item *MenuItem) Show() error {
	return showMenuItem(item)
}

// Checked returns if the menu item has a check mark
func (item *MenuItem) Checked() bool {
	return item.checked
}

// Check a menu item regardless if it's previously checked or not
func (item *MenuItem) Check() error {
	item.checked = true
	return item.update()
}

// Uncheck a menu item regardless if it's previously unchecked or not
func (item *MenuItem) Uncheck() error {
	item.checked = false
	return item.update()
}

// update propogates changes on a menu item to systray
func (item *MenuItem) update() error {
	menuItemsLock.Lock()
	defer menuItemsLock.Unlock()
	menuItems[item.id] = item
	return addOrUpdateMenuItem(item)
}

func systrayMenuItemSelected(id int32) {
	menuItemsLock.RLock()
	item := menuItems[id]
	menuItemsLock.RUnlock()
	select {
	case item.ClickedCh <- struct{}{}:
		// in case no one waiting for the channel
	default:
	}
}
