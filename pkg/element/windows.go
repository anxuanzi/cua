//go:build windows

// Package element provides cross-platform UI element access via accessibility APIs.
// This file contains the Windows implementation using UI Automation.
package element

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

func init() {
	// Set the finder implementation constructor for Windows
	newFinderImpl = newWindowsFinder

	// Set platform-specific function implementations
	focusElement = windowsFocusElement
	performAction = windowsPerformAction
	setValue = windowsSetValue
	loadChildren = windowsLoadChildren
}

// Windows COM GUIDs
var (
	CLSID_CUIAutomation = &GUID{0xff48dba4, 0x60ef, 0x4201, [8]byte{0xaa, 0x87, 0x54, 0x10, 0x3e, 0xef, 0x59, 0x4e}}
	IID_IUIAutomation   = &GUID{0x30cbe57d, 0xd9d0, 0x452a, [8]byte{0xab, 0x13, 0x7a, 0xc5, 0xac, 0x48, 0x25, 0xee}}
)

// GUID represents a Windows GUID
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// COM interface offsets (vtable indices)
const (
	queryInterfaceOffset = 0
	addRefOffset         = 1
	releaseOffset        = 2

	// IUIAutomation methods (after IUnknown)
	compareElementsOffset             = 3
	compareRuntimeIdsOffset           = 4
	getRootElementOffset              = 5
	elementFromHandleOffset           = 6
	elementFromPointOffset            = 7
	getFocusedElementOffset           = 8
	getRootElementBuildCacheOffset    = 9
	elementFromHandleBuildCacheOffset = 10
	elementFromPointBuildCacheOffset  = 11
	getFocusedElementBuildCacheOffset = 12
	createTreeWalkerOffset            = 13
	controlViewWalkerOffset           = 14
	contentViewWalkerOffset           = 15
	rawViewWalkerOffset               = 16
	rawViewConditionOffset            = 17
	controlViewConditionOffset        = 18
	contentViewConditionOffset        = 19
	createCacheRequestOffset          = 20
	createTrueConditionOffset         = 21
	createFalseConditionOffset        = 22
	createPropertyConditionOffset     = 23
	createPropertyConditionExOffset   = 24
	createAndConditionOffset          = 25
	createAndConditionFromArrayOffset = 26
	createOrConditionOffset           = 27
	createOrConditionFromArrayOffset  = 28
	createNotConditionOffset          = 29

	// IUIAutomationElement methods (after IUnknown)
	setFocusOffset                      = 3
	getRuntimeIdOffset                  = 4
	findFirstOffset                     = 5
	findAllOffset                       = 6
	findFirstBuildCacheOffset           = 7
	findAllBuildCacheOffset             = 8
	buildUpdatedCacheOffset             = 9
	getCurrentPropertyValueOffset       = 10
	getCurrentPropertyValueExOffset     = 11
	getCachedPropertyValueOffset        = 12
	getCachedPropertyValueExOffset      = 13
	getCurrentPatternAsOffset           = 14
	getCachedPatternAsOffset            = 15
	getCurrentPatternOffset             = 16
	getCachedPatternOffset              = 17
	getCurrentParentOffset              = 18
	getCurrentFirstChildOffset          = 19
	getCurrentLastChildOffset           = 20
	getCurrentNextSiblingOffset         = 21
	getCurrentPreviousSiblingOffset     = 22
	getCurrentProcessIdOffset           = 23
	getCurrentControlTypeOffset         = 24
	getCurrentLocalizControlTypeOffset  = 25
	getCurrentNameOffset                = 26
	getCurrentAcceleratorKeyOffset      = 27
	getCurrentAccessKeyOffset           = 28
	getCurrentHasKeyboardFocusOffset    = 29
	getCurrentIsKeyboardFocusableOffset = 30
	getCurrentIsEnabledOffset           = 31
	getCurrentAutomationIdOffset        = 32
	getCurrentClassNameOffset           = 33
	getCurrentHelpTextOffset            = 34
	getCurrentCultureOffset             = 35
	getCurrentIsControlElementOffset    = 36
	getCurrentIsContentElementOffset    = 37
	getCurrentIsPasswordOffset          = 38
	getCurrentNativeWindowHandleOffset  = 39
	getCurrentItemTypeOffset            = 40
	getCurrentIsOffscreenOffset         = 41
	getCurrentOrientationOffset         = 42
	getCurrentFrameworkIdOffset         = 43
	getCurrentIsRequiredForFormOffset   = 44
	getCurrentItemStatusOffset          = 45
	getCurrentBoundingRectangleOffset   = 46
	getCurrentLabeledByOffset           = 47
	getCurrentAriaRoleOffset            = 48
	getCurrentAriaPropertiesOffset      = 49
	getCurrentIsDataValidForFormOffset  = 50
	getCurrentControllerForOffset       = 51
	getCurrentDescribedByOffset         = 52
	getCurrentFlowsToOffset             = 53
	getCurrentProviderDescriptionOffset = 54

	// IUIAutomationTreeWalker methods (after IUnknown)
	getParentElementOffset          = 3
	getFirstChildElementOffset      = 4
	getLastChildElementOffset       = 5
	getNextSiblingElementOffset     = 6
	getPreviousSiblingElementOffset = 7
)

// UI Automation Control Type IDs
const (
	UIA_ButtonControlTypeId       = 50000
	UIA_CalendarControlTypeId     = 50001
	UIA_CheckBoxControlTypeId     = 50002
	UIA_ComboBoxControlTypeId     = 50003
	UIA_EditControlTypeId         = 50004
	UIA_HyperlinkControlTypeId    = 50005
	UIA_ImageControlTypeId        = 50006
	UIA_ListItemControlTypeId     = 50007
	UIA_ListControlTypeId         = 50008
	UIA_MenuControlTypeId         = 50009
	UIA_MenuBarControlTypeId      = 50010
	UIA_MenuItemControlTypeId     = 50011
	UIA_ProgressBarControlTypeId  = 50012
	UIA_RadioButtonControlTypeId  = 50013
	UIA_ScrollBarControlTypeId    = 50014
	UIA_SliderControlTypeId       = 50015
	UIA_SpinnerControlTypeId      = 50016
	UIA_StatusBarControlTypeId    = 50017
	UIA_TabControlTypeId          = 50018
	UIA_TabItemControlTypeId      = 50019
	UIA_TextControlTypeId         = 50020
	UIA_ToolBarControlTypeId      = 50021
	UIA_ToolTipControlTypeId      = 50022
	UIA_TreeControlTypeId         = 50023
	UIA_TreeItemControlTypeId     = 50024
	UIA_CustomControlTypeId       = 50025
	UIA_GroupControlTypeId        = 50026
	UIA_ThumbControlTypeId        = 50027
	UIA_DataGridControlTypeId     = 50028
	UIA_DataItemControlTypeId     = 50029
	UIA_DocumentControlTypeId     = 50030
	UIA_SplitButtonControlTypeId  = 50031
	UIA_WindowControlTypeId       = 50032
	UIA_PaneControlTypeId         = 50033
	UIA_HeaderControlTypeId       = 50034
	UIA_HeaderItemControlTypeId   = 50035
	UIA_TableControlTypeId        = 50036
	UIA_TitleBarControlTypeId     = 50037
	UIA_SeparatorControlTypeId    = 50038
	UIA_SemanticZoomControlTypeId = 50039
	UIA_AppBarControlTypeId       = 50040
)

// TreeScope constants
const (
	TreeScope_None        = 0
	TreeScope_Element     = 1
	TreeScope_Children    = 2
	TreeScope_Descendants = 4
	TreeScope_Parent      = 8
	TreeScope_Ancestors   = 16
	TreeScope_Subtree     = TreeScope_Element | TreeScope_Descendants
)

// Windows API DLLs
var (
	ole32    = syscall.NewLazyDLL("ole32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	oleaut32 = syscall.NewLazyDLL("oleaut32.dll")

	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoUninitialize   = ole32.NewProc("CoUninitialize")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")

	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")

	procSysFreeString = oleaut32.NewProc("SysFreeString")
)

// COM threading model
const (
	COINIT_APARTMENTTHREADED = 0x2
	COINIT_MULTITHREADED     = 0x0
	CLSCTX_INPROC_SERVER     = 0x1
)

// RECT structure for Windows
type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// windowsFinder implements finderImpl for Windows using UI Automation
type windowsFinder struct {
	automation     uintptr // IUIAutomation*
	treeWalker     uintptr // IUIAutomationTreeWalker*
	mu             sync.Mutex
	comInitialized bool
}

func newWindowsFinder() (finderImpl, error) {
	f := &windowsFinder{}

	// Initialize COM on this thread
	runtime.LockOSThread()
	hr, _, _ := procCoInitializeEx.Call(0, COINIT_MULTITHREADED)
	if hr != 0 && hr != 1 { // S_OK or S_FALSE (already initialized)
		return nil, fmt.Errorf("CoInitializeEx failed: 0x%x", hr)
	}
	f.comInitialized = true

	// Create IUIAutomation instance
	var automation uintptr
	hr, _, _ = procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(CLSID_CUIAutomation)),
		0,
		CLSCTX_INPROC_SERVER,
		uintptr(unsafe.Pointer(IID_IUIAutomation)),
		uintptr(unsafe.Pointer(&automation)),
	)
	if hr != 0 {
		procCoUninitialize.Call()
		return nil, fmt.Errorf("CoCreateInstance failed: 0x%x", hr)
	}
	f.automation = automation

	// Get the raw view tree walker for traversal
	var walker uintptr
	vtbl := *(*uintptr)(unsafe.Pointer(automation))
	hr, _, _ = syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + rawViewWalkerOffset*unsafe.Sizeof(uintptr(0)))),
		automation,
		uintptr(unsafe.Pointer(&walker)),
	)
	if hr == 0 {
		f.treeWalker = walker
	}

	return f, nil
}

func (f *windowsFinder) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.treeWalker != 0 {
		release(f.treeWalker)
		f.treeWalker = 0
	}

	if f.automation != 0 {
		release(f.automation)
		f.automation = 0
	}

	if f.comInitialized {
		procCoUninitialize.Call()
		f.comInitialized = false
		runtime.UnlockOSThread()
	}

	return nil
}

func (f *windowsFinder) Root() (*Element, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.automation == 0 {
		return nil, ErrInvalidElement
	}

	var root uintptr
	vtbl := *(*uintptr)(unsafe.Pointer(f.automation))
	hr, _, _ := syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getRootElementOffset*unsafe.Sizeof(uintptr(0)))),
		f.automation,
		uintptr(unsafe.Pointer(&root)),
	)
	if hr != 0 || root == 0 {
		return nil, fmt.Errorf("GetRootElement failed: 0x%x", hr)
	}

	return f.elementFromUIAElement(root, nil)
}

func (f *windowsFinder) FocusedApplication() (*Element, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Get the foreground window
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return nil, ErrNotFound
	}

	// Get PID of foreground window
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	// Get element from window handle
	if f.automation == 0 {
		return nil, ErrInvalidElement
	}

	var element uintptr
	vtbl := *(*uintptr)(unsafe.Pointer(f.automation))
	hr, _, _ := syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + elementFromHandleOffset*unsafe.Sizeof(uintptr(0)))),
		f.automation,
		hwnd,
		uintptr(unsafe.Pointer(&element)),
	)
	if hr != 0 || element == 0 {
		return nil, fmt.Errorf("ElementFromHandle failed: 0x%x", hr)
	}

	return f.elementFromUIAElement(element, nil)
}

func (f *windowsFinder) FocusedElement() (*Element, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.automation == 0 {
		return nil, ErrInvalidElement
	}

	var focused uintptr
	vtbl := *(*uintptr)(unsafe.Pointer(f.automation))
	hr, _, _ := syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getFocusedElementOffset*unsafe.Sizeof(uintptr(0)))),
		f.automation,
		uintptr(unsafe.Pointer(&focused)),
	)
	if hr != 0 || focused == 0 {
		return nil, ErrNoFocus
	}

	return f.elementFromUIAElement(focused, nil)
}

func (f *windowsFinder) ApplicationByPID(pid int) (*Element, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Find windows belonging to this PID
	hwnd := findWindowByPID(uint32(pid))
	if hwnd == 0 {
		return nil, ErrNotFound
	}

	if f.automation == 0 {
		return nil, ErrInvalidElement
	}

	var element uintptr
	vtbl := *(*uintptr)(unsafe.Pointer(f.automation))
	hr, _, _ := syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + elementFromHandleOffset*unsafe.Sizeof(uintptr(0)))),
		f.automation,
		hwnd,
		uintptr(unsafe.Pointer(&element)),
	)
	if hr != 0 || element == 0 {
		return nil, fmt.Errorf("ElementFromHandle failed: 0x%x", hr)
	}

	return f.elementFromUIAElement(element, nil)
}

func (f *windowsFinder) ApplicationByName(name string) (*Element, error) {
	apps, err := f.AllApplications()
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)
	for _, app := range apps {
		if strings.ToLower(app.Name) == nameLower || strings.ToLower(app.Title) == nameLower {
			return app, nil
		}
	}
	return nil, ErrNotFound
}

func (f *windowsFinder) AllApplications() ([]*Element, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.automation == 0 {
		return nil, ErrInvalidElement
	}

	// Get all visible top-level windows
	hwnds := findAllTopLevelWindows()

	var apps []*Element
	vtbl := *(*uintptr)(unsafe.Pointer(f.automation))

	for _, hwnd := range hwnds {
		var element uintptr
		hr, _, _ := syscall.SyscallN(
			*(*uintptr)(unsafe.Pointer(vtbl + elementFromHandleOffset*unsafe.Sizeof(uintptr(0)))),
			f.automation,
			hwnd,
			uintptr(unsafe.Pointer(&element)),
		)
		if hr == 0 && element != 0 {
			elem, err := f.elementFromUIAElement(element, nil)
			if err == nil {
				apps = append(apps, elem)
			}
		}
	}

	return apps, nil
}

// elementFromUIAElement creates an Element from a UI Automation element
func (f *windowsFinder) elementFromUIAElement(uiaElement uintptr, parent *Element) (*Element, error) {
	if uiaElement == 0 {
		return nil, ErrInvalidElement
	}

	elem := &Element{
		Parent:     parent,
		Attributes: make(map[string]interface{}),
		handle:     uiaElement,
		Enabled:    true, // default
	}

	vtbl := *(*uintptr)(unsafe.Pointer(uiaElement))

	// Get control type
	var controlType int32
	hr, _, _ := syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getCurrentControlTypeOffset*unsafe.Sizeof(uintptr(0)))),
		uiaElement,
		uintptr(unsafe.Pointer(&controlType)),
	)
	if hr == 0 {
		elem.Role = mapWindowsControlType(controlType)
	}

	// Get name
	var bstrName uintptr
	hr, _, _ = syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getCurrentNameOffset*unsafe.Sizeof(uintptr(0)))),
		uiaElement,
		uintptr(unsafe.Pointer(&bstrName)),
	)
	if hr == 0 && bstrName != 0 {
		elem.Name = bstrToString(bstrName)
		elem.Title = elem.Name
		procSysFreeString.Call(bstrName)
	}

	// Get bounding rectangle
	var rect RECT
	hr, _, _ = syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getCurrentBoundingRectangleOffset*unsafe.Sizeof(uintptr(0)))),
		uiaElement,
		uintptr(unsafe.Pointer(&rect)),
	)
	if hr == 0 {
		elem.Bounds = Rect{
			X:      int(rect.Left),
			Y:      int(rect.Top),
			Width:  int(rect.Right - rect.Left),
			Height: int(rect.Bottom - rect.Top),
		}
	}

	// Get enabled state
	var isEnabled int32
	hr, _, _ = syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getCurrentIsEnabledOffset*unsafe.Sizeof(uintptr(0)))),
		uiaElement,
		uintptr(unsafe.Pointer(&isEnabled)),
	)
	if hr == 0 {
		elem.Enabled = isEnabled != 0
	}

	// Get focus state
	var hasFocus int32
	hr, _, _ = syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getCurrentHasKeyboardFocusOffset*unsafe.Sizeof(uintptr(0)))),
		uiaElement,
		uintptr(unsafe.Pointer(&hasFocus)),
	)
	if hr == 0 {
		elem.Focused = hasFocus != 0
	}

	// Get process ID
	var pid int32
	hr, _, _ = syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getCurrentProcessIdOffset*unsafe.Sizeof(uintptr(0)))),
		uiaElement,
		uintptr(unsafe.Pointer(&pid)),
	)
	if hr == 0 {
		elem.PID = int(pid)
	}

	// Generate ID
	elem.ID = fmt.Sprintf("%d-%x", elem.PID, uiaElement)

	// Set finalizer to release the COM object
	runtime.SetFinalizer(elem, func(e *Element) {
		if e.handle != nil {
			if ptr, ok := e.handle.(uintptr); ok && ptr != 0 {
				release(ptr)
			}
		}
	})

	return elem, nil
}

// release calls IUnknown::Release on a COM object
func release(obj uintptr) {
	if obj == 0 {
		return
	}
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + releaseOffset*unsafe.Sizeof(uintptr(0)))),
		obj,
	)
}

// bstrToString converts a BSTR to a Go string
func bstrToString(bstr uintptr) string {
	if bstr == 0 {
		return ""
	}

	// BSTR has length prefix at -4 offset (in bytes)
	length := *(*uint32)(unsafe.Pointer(bstr - 4))
	if length == 0 {
		return ""
	}

	// BSTR is UTF-16 encoded
	chars := length / 2
	utf16 := make([]uint16, chars)
	for i := uint32(0); i < chars; i++ {
		utf16[i] = *(*uint16)(unsafe.Pointer(bstr + uintptr(i*2)))
	}

	return syscall.UTF16ToString(utf16)
}

// mapWindowsControlType converts a Windows control type ID to our Role type
func mapWindowsControlType(controlType int32) Role {
	switch controlType {
	case UIA_ButtonControlTypeId:
		return RoleButton
	case UIA_CalendarControlTypeId:
		return RoleGroup
	case UIA_CheckBoxControlTypeId:
		return RoleCheckbox
	case UIA_ComboBoxControlTypeId:
		return RoleComboBox
	case UIA_EditControlTypeId:
		return RoleTextField
	case UIA_HyperlinkControlTypeId:
		return RoleLink
	case UIA_ImageControlTypeId:
		return RoleImage
	case UIA_ListItemControlTypeId:
		return RoleListItem
	case UIA_ListControlTypeId:
		return RoleList
	case UIA_MenuControlTypeId:
		return RoleMenu
	case UIA_MenuBarControlTypeId:
		return RoleMenuBar
	case UIA_MenuItemControlTypeId:
		return RoleMenuItem
	case UIA_ProgressBarControlTypeId:
		return RoleProgressBar
	case UIA_RadioButtonControlTypeId:
		return RoleRadioButton
	case UIA_ScrollBarControlTypeId:
		return RoleScrollBar
	case UIA_SliderControlTypeId:
		return RoleSlider
	case UIA_TabControlTypeId:
		return RoleTabGroup
	case UIA_TabItemControlTypeId:
		return RoleTab
	case UIA_TextControlTypeId:
		return RoleStaticText
	case UIA_ToolBarControlTypeId:
		return RoleToolbar
	case UIA_TreeControlTypeId:
		return RoleList
	case UIA_TreeItemControlTypeId:
		return RoleListItem
	case UIA_GroupControlTypeId:
		return RoleGroup
	case UIA_DataGridControlTypeId:
		return RoleTable
	case UIA_DataItemControlTypeId:
		return RoleRow
	case UIA_DocumentControlTypeId:
		return RoleTextArea
	case UIA_WindowControlTypeId:
		return RoleWindow
	case UIA_PaneControlTypeId:
		return RoleGroup
	case UIA_TableControlTypeId:
		return RoleTable
	case UIA_SeparatorControlTypeId:
		return RoleSplitter
	default:
		return RoleUnknown
	}
}

// findWindowByPID finds the main window for a process ID
func findWindowByPID(targetPID uint32) uintptr {
	var result uintptr

	callback := syscall.NewCallback(func(hwnd, lparam uintptr) uintptr {
		var pid uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		if pid == targetPID {
			// Check if window is visible
			visible, _, _ := procIsWindowVisible.Call(hwnd)
			if visible != 0 {
				result = hwnd
				return 0 // stop enumeration
			}
		}
		return 1 // continue
	})

	procEnumWindows.Call(callback, 0)
	return result
}

// findAllTopLevelWindows returns all visible top-level windows
func findAllTopLevelWindows() []uintptr {
	var hwnds []uintptr

	callback := syscall.NewCallback(func(hwnd, lparam uintptr) uintptr {
		visible, _, _ := procIsWindowVisible.Call(hwnd)
		if visible != 0 {
			// Check if window has a title (filter out hidden/system windows)
			buf := make([]uint16, 256)
			length, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
			if length > 0 {
				hwnds = append(hwnds, hwnd)
			}
		}
		return 1 // continue
	})

	procEnumWindows.Call(callback, 0)
	return hwnds
}

// Platform-specific action implementations

func windowsFocusElement(e *Element) error {
	if e.handle == nil {
		return ErrInvalidElement
	}

	ptr, ok := e.handle.(uintptr)
	if !ok || ptr == 0 {
		return ErrInvalidElement
	}

	vtbl := *(*uintptr)(unsafe.Pointer(ptr))
	hr, _, _ := syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + setFocusOffset*unsafe.Sizeof(uintptr(0)))),
		ptr,
	)
	if hr != 0 {
		return fmt.Errorf("SetFocus failed: 0x%x", hr)
	}
	return nil
}

func windowsPerformAction(e *Element, action string) error {
	// On Windows, actions are performed via patterns (InvokePattern, etc.)
	// For now, only support basic invoke
	if action == "AXPress" || action == "invoke" {
		return windowsInvoke(e)
	}
	return ErrNotSupported
}

func windowsInvoke(e *Element) error {
	if e.handle == nil {
		return ErrInvalidElement
	}

	ptr, ok := e.handle.(uintptr)
	if !ok || ptr == 0 {
		return ErrInvalidElement
	}

	// Get InvokePattern (pattern ID = 10000)
	const UIA_InvokePatternId = 10000

	var pattern uintptr
	vtbl := *(*uintptr)(unsafe.Pointer(ptr))
	hr, _, _ := syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getCurrentPatternOffset*unsafe.Sizeof(uintptr(0)))),
		ptr,
		UIA_InvokePatternId,
		uintptr(unsafe.Pointer(&pattern)),
	)
	if hr != 0 || pattern == 0 {
		return fmt.Errorf("GetCurrentPattern failed: 0x%x", hr)
	}
	defer release(pattern)

	// Call Invoke (method index 3 after IUnknown)
	patternVtbl := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ = syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(patternVtbl + 3*unsafe.Sizeof(uintptr(0)))),
		pattern,
	)
	if hr != 0 {
		return fmt.Errorf("Invoke failed: 0x%x", hr)
	}
	return nil
}

func windowsSetValue(e *Element, value string) error {
	if e.handle == nil {
		return ErrInvalidElement
	}

	ptr, ok := e.handle.(uintptr)
	if !ok || ptr == 0 {
		return ErrInvalidElement
	}

	// Get ValuePattern (pattern ID = 10002)
	const UIA_ValuePatternId = 10002

	var pattern uintptr
	vtbl := *(*uintptr)(unsafe.Pointer(ptr))
	hr, _, _ := syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(vtbl + getCurrentPatternOffset*unsafe.Sizeof(uintptr(0)))),
		ptr,
		UIA_ValuePatternId,
		uintptr(unsafe.Pointer(&pattern)),
	)
	if hr != 0 || pattern == 0 {
		return fmt.Errorf("GetCurrentPattern failed: 0x%x", hr)
	}
	defer release(pattern)

	// Convert string to UTF-16 for Windows
	utf16Value, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		return err
	}

	// Call SetValue (method index 3 after IUnknown for IUIAutomationValuePattern)
	patternVtbl := *(*uintptr)(unsafe.Pointer(pattern))
	hr, _, _ = syscall.SyscallN(
		*(*uintptr)(unsafe.Pointer(patternVtbl + 3*unsafe.Sizeof(uintptr(0)))),
		pattern,
		uintptr(unsafe.Pointer(utf16Value)),
	)
	if hr != 0 {
		return fmt.Errorf("SetValue failed: 0x%x", hr)
	}
	return nil
}

func windowsLoadChildren(e *Element) error {
	if e.handle == nil {
		return ErrInvalidElement
	}

	ptr, ok := e.handle.(uintptr)
	if !ok || ptr == 0 {
		return ErrInvalidElement
	}

	// We need access to the finder's tree walker
	// For now, use a simplified approach with FindAll

	// Get a true condition for finding all children
	// This is a simplified implementation - a full implementation would
	// use the tree walker from the finder

	e.Children = []*Element{}

	// Use the raw view tree walker if we have access to it
	// For elements created through the finder, we could add a reference
	// to the finder to enable proper tree walking

	return nil
}
