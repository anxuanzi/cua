//go:build darwin

package element

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Foundation -framework AppKit

#include <ApplicationServices/ApplicationServices.h>
#include <Foundation/Foundation.h>
#include <AppKit/AppKit.h>

// Check if accessibility is enabled
static int ax_is_trusted() {
    return AXIsProcessTrusted();
}

// Check if accessibility is enabled with prompt option
static int ax_is_trusted_with_prompt() {
    NSDictionary *options = @{(__bridge NSString *)kAXTrustedCheckOptionPrompt: @YES};
    return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
}

// Create system-wide element
static AXUIElementRef ax_create_system_wide() {
    return AXUIElementCreateSystemWide();
}

// Create application element by PID
static AXUIElementRef ax_create_application(int pid) {
    return AXUIElementCreateApplication(pid);
}

// Get attribute value
static CFTypeRef ax_copy_attribute_value(AXUIElementRef element, CFStringRef attribute) {
    CFTypeRef value = NULL;
    AXError err = AXUIElementCopyAttributeValue(element, attribute, &value);
    if (err != kAXErrorSuccess) {
        return NULL;
    }
    return value;
}

// Get attribute names
static CFArrayRef ax_copy_attribute_names(AXUIElementRef element) {
    CFArrayRef names = NULL;
    AXError err = AXUIElementCopyAttributeNames(element, &names);
    if (err != kAXErrorSuccess) {
        return NULL;
    }
    return names;
}

// Get actions
static CFArrayRef ax_copy_action_names(AXUIElementRef element) {
    CFArrayRef names = NULL;
    AXError err = AXUIElementCopyActionNames(element, &names);
    if (err != kAXErrorSuccess) {
        return NULL;
    }
    return names;
}

// Perform action
static int ax_perform_action(AXUIElementRef element, CFStringRef action) {
    AXError err = AXUIElementPerformAction(element, action);
    return err == kAXErrorSuccess ? 0 : (int)err;
}

// Set attribute value
static int ax_set_attribute_value(AXUIElementRef element, CFStringRef attribute, CFTypeRef value) {
    AXError err = AXUIElementSetAttributeValue(element, attribute, value);
    return err == kAXErrorSuccess ? 0 : (int)err;
}

// Get PID from element
static int ax_get_pid(AXUIElementRef element) {
    pid_t pid = 0;
    AXError err = AXUIElementGetPid(element, &pid);
    if (err != kAXErrorSuccess) {
        return -1;
    }
    return (int)pid;
}

// Get element at position
static AXUIElementRef ax_copy_element_at_position(AXUIElementRef parent, float x, float y) {
    AXUIElementRef element = NULL;
    AXError err = AXUIElementCopyElementAtPosition(parent, x, y, &element);
    if (err != kAXErrorSuccess) {
        return NULL;
    }
    return element;
}

// Get frontmost app PID
static int ax_get_frontmost_app_pid() {
    NSRunningApplication *frontApp = [[NSWorkspace sharedWorkspace] frontmostApplication];
    if (frontApp == nil) {
        return -1;
    }
    return (int)[frontApp processIdentifier];
}

// Get all running app PIDs
static void ax_get_running_apps(int *pids, int *count, int maxCount) {
    NSArray<NSRunningApplication *> *apps = [[NSWorkspace sharedWorkspace] runningApplications];
    int i = 0;
    for (NSRunningApplication *app in apps) {
        if (i >= maxCount) break;
        // Only include regular apps (not background agents)
        if (app.activationPolicy == NSApplicationActivationPolicyRegular) {
            pids[i++] = (int)[app processIdentifier];
        }
    }
    *count = i;
}

// Get CFString as C string (caller must free)
static char* cf_string_to_cstring(CFStringRef str) {
    if (str == NULL) return NULL;

    CFIndex length = CFStringGetLength(str);
    CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
    char *buffer = (char *)malloc(maxSize);
    if (buffer == NULL) return NULL;

    if (!CFStringGetCString(str, buffer, maxSize, kCFStringEncodingUTF8)) {
        free(buffer);
        return NULL;
    }
    return buffer;
}

// Create CFString from C string
static CFStringRef cstring_to_cf_string(const char *str) {
    return CFStringCreateWithCString(kCFAllocatorDefault, str, kCFStringEncodingUTF8);
}

// Get position from AXValue
static int ax_value_get_point(AXValueRef value, float *x, float *y) {
    CGPoint point;
    if (AXValueGetValue(value, kAXValueCGPointType, &point)) {
        *x = point.x;
        *y = point.y;
        return 1;
    }
    return 0;
}

// Get size from AXValue
static int ax_value_get_size(AXValueRef value, float *width, float *height) {
    CGSize size;
    if (AXValueGetValue(value, kAXValueCGSizeType, &size)) {
        *width = size.width;
        *height = size.height;
        return 1;
    }
    return 0;
}

*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

func init() {
	// Set platform-specific function implementations
	focusElement = darwinFocusElement
	performAction = darwinPerformAction
	setValue = darwinSetValue
	loadChildren = darwinLoadChildren

	// Set the finder implementation constructor
	newFinderImpl = newDarwinFinder
}

// darwinFinder implements finderImpl for macOS
type darwinFinder struct {
	systemWide C.AXUIElementRef
	mu         sync.Mutex
}

func newDarwinFinder() (finderImpl, error) {
	// Check accessibility permission
	if C.ax_is_trusted() == 0 {
		// Prompt the user
		C.ax_is_trusted_with_prompt()
		return nil, ErrPermissionDenied
	}

	systemWide := C.ax_create_system_wide()
	if systemWide == 0 {
		return nil, fmt.Errorf("failed to create system-wide accessibility element")
	}

	return &darwinFinder{
		systemWide: systemWide,
	}, nil
}

func (f *darwinFinder) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.systemWide != 0 {
		C.CFRelease(C.CFTypeRef(f.systemWide))
		f.systemWide = 0
	}
	return nil
}

func (f *darwinFinder) Root() (*Element, error) {
	return f.elementFromRef(f.systemWide, nil)
}

func (f *darwinFinder) FocusedApplication() (*Element, error) {
	pid := int(C.ax_get_frontmost_app_pid())
	if pid < 0 {
		return nil, fmt.Errorf("failed to get frontmost application")
	}
	return f.ApplicationByPID(pid)
}

func (f *darwinFinder) FocusedElement() (*Element, error) {
	app, err := f.FocusedApplication()
	if err != nil {
		return nil, err
	}

	ref := app.handle.(C.AXUIElementRef)
	attrName := C.cstring_to_cf_string(C.CString("AXFocusedUIElement"))
	defer C.CFRelease(C.CFTypeRef(attrName))

	value := C.ax_copy_attribute_value(ref, attrName)
	if value == 0 {
		return nil, ErrNoFocus
	}
	defer C.CFRelease(value)

	focusedRef := C.AXUIElementRef(value)
	return f.elementFromRef(focusedRef, nil)
}

func (f *darwinFinder) ApplicationByPID(pid int) (*Element, error) {
	ref := C.ax_create_application(C.int(pid))
	if ref == 0 {
		return nil, fmt.Errorf("failed to get application with PID %d", pid)
	}

	elem, err := f.elementFromRef(ref, nil)
	if err != nil {
		C.CFRelease(C.CFTypeRef(ref))
		return nil, err
	}
	return elem, nil
}

func (f *darwinFinder) ApplicationByName(name string) (*Element, error) {
	apps, err := f.AllApplications()
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.Name == name || app.Title == name {
			return app, nil
		}
	}
	return nil, ErrNotFound
}

func (f *darwinFinder) AllApplications() ([]*Element, error) {
	const maxApps = 100
	pids := make([]C.int, maxApps)
	var count C.int

	C.ax_get_running_apps(&pids[0], &count, C.int(maxApps))

	var apps []*Element
	for i := 0; i < int(count); i++ {
		pid := int(pids[i])
		app, err := f.ApplicationByPID(pid)
		if err == nil && app != nil {
			apps = append(apps, app)
		}
	}

	return apps, nil
}

// elementFromRef creates an Element from an AXUIElementRef
func (f *darwinFinder) elementFromRef(ref C.AXUIElementRef, parent *Element) (*Element, error) {
	if ref == 0 {
		return nil, ErrInvalidElement
	}

	elem := &Element{
		Parent:     parent,
		Attributes: make(map[string]interface{}),
		handle:     ref,
	}

	// Get PID
	elem.PID = int(C.ax_get_pid(ref))

	// Get role
	elem.Role = mapRole(f.getStringAttribute(ref, "AXRole"))

	// Get name/title
	elem.Name = f.getStringAttribute(ref, "AXTitle")
	if elem.Name == "" {
		elem.Name = f.getStringAttribute(ref, "AXDescription")
	}
	elem.Title = f.getStringAttribute(ref, "AXTitle")
	elem.Description = f.getStringAttribute(ref, "AXDescription")
	elem.Value = f.getStringAttribute(ref, "AXValue")

	// Get enabled state
	elem.Enabled = f.getBoolAttribute(ref, "AXEnabled")

	// Get focused state
	elem.Focused = f.getBoolAttribute(ref, "AXFocused")

	// Get selected state
	elem.Selected = f.getBoolAttribute(ref, "AXSelected")

	// Get bounds (position + size)
	x, y := f.getPositionAttribute(ref, "AXPosition")
	w, h := f.getSizeAttribute(ref, "AXSize")
	elem.Bounds = Rect{
		X:      int(x),
		Y:      int(y),
		Width:  int(w),
		Height: int(h),
	}

	// Generate a unique ID
	elem.ID = fmt.Sprintf("%d-%x", elem.PID, uintptr(ref))

	// Set finalizer to release the ref when element is garbage collected
	runtime.SetFinalizer(elem, func(e *Element) {
		if e.handle != nil {
			C.CFRelease(C.CFTypeRef(e.handle.(C.AXUIElementRef)))
		}
	})

	return elem, nil
}

func (f *darwinFinder) getStringAttribute(ref C.AXUIElementRef, name string) string {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	attrName := C.cstring_to_cf_string(cName)
	defer C.CFRelease(C.CFTypeRef(attrName))

	value := C.ax_copy_attribute_value(ref, attrName)
	if value == 0 {
		return ""
	}
	defer C.CFRelease(value)

	// Check if it's a string
	typeID := C.CFGetTypeID(value)
	if typeID != C.CFStringGetTypeID() {
		return ""
	}

	cStr := C.cf_string_to_cstring(C.CFStringRef(value))
	if cStr == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cStr))

	return C.GoString(cStr)
}

func (f *darwinFinder) getBoolAttribute(ref C.AXUIElementRef, name string) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	attrName := C.cstring_to_cf_string(cName)
	defer C.CFRelease(C.CFTypeRef(attrName))

	value := C.ax_copy_attribute_value(ref, attrName)
	if value == 0 {
		return false
	}
	defer C.CFRelease(value)

	typeID := C.CFGetTypeID(value)
	if typeID == C.CFBooleanGetTypeID() {
		return C.CFBooleanGetValue(C.CFBooleanRef(value)) != 0
	}

	return false
}

func (f *darwinFinder) getPositionAttribute(ref C.AXUIElementRef, name string) (float32, float32) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	attrName := C.cstring_to_cf_string(cName)
	defer C.CFRelease(C.CFTypeRef(attrName))

	value := C.ax_copy_attribute_value(ref, attrName)
	if value == 0 {
		return 0, 0
	}
	defer C.CFRelease(value)

	var x, y C.float
	axValue := C.AXValueRef(unsafe.Pointer(value))
	if C.ax_value_get_point(axValue, &x, &y) != 0 {
		return float32(x), float32(y)
	}
	return 0, 0
}

func (f *darwinFinder) getSizeAttribute(ref C.AXUIElementRef, name string) (float32, float32) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	attrName := C.cstring_to_cf_string(cName)
	defer C.CFRelease(C.CFTypeRef(attrName))

	value := C.ax_copy_attribute_value(ref, attrName)
	if value == 0 {
		return 0, 0
	}
	defer C.CFRelease(value)

	var w, h C.float
	axValue := C.AXValueRef(unsafe.Pointer(value))
	if C.ax_value_get_size(axValue, &w, &h) != 0 {
		return float32(w), float32(h)
	}
	return 0, 0
}

// Platform-specific action implementations

func darwinFocusElement(e *Element) error {
	if e.handle == nil {
		return ErrInvalidElement
	}

	ref := e.handle.(C.AXUIElementRef)

	// Try to raise the window first
	actionName := C.cstring_to_cf_string(C.CString("AXRaise"))
	defer C.CFRelease(C.CFTypeRef(actionName))
	C.ax_perform_action(ref, actionName)

	// Then set focus
	attrName := C.cstring_to_cf_string(C.CString("AXFocused"))
	defer C.CFRelease(C.CFTypeRef(attrName))

	trueVal := C.kCFBooleanTrue
	result := C.ax_set_attribute_value(ref, attrName, C.CFTypeRef(trueVal))
	if result != 0 {
		return fmt.Errorf("failed to focus element: error %d", result)
	}
	return nil
}

func darwinPerformAction(e *Element, action string) error {
	if e.handle == nil {
		return ErrInvalidElement
	}

	ref := e.handle.(C.AXUIElementRef)

	cAction := C.CString(action)
	defer C.free(unsafe.Pointer(cAction))

	actionName := C.cstring_to_cf_string(cAction)
	defer C.CFRelease(C.CFTypeRef(actionName))

	result := C.ax_perform_action(ref, actionName)
	if result != 0 {
		return fmt.Errorf("failed to perform action %s: error %d", action, result)
	}
	return nil
}

func darwinSetValue(e *Element, value string) error {
	if e.handle == nil {
		return ErrInvalidElement
	}

	ref := e.handle.(C.AXUIElementRef)

	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	attrName := C.cstring_to_cf_string(C.CString("AXValue"))
	defer C.CFRelease(C.CFTypeRef(attrName))

	cfValue := C.cstring_to_cf_string(cValue)
	defer C.CFRelease(C.CFTypeRef(cfValue))

	result := C.ax_set_attribute_value(ref, attrName, C.CFTypeRef(cfValue))
	if result != 0 {
		return fmt.Errorf("failed to set value: error %d", result)
	}
	return nil
}

func darwinLoadChildren(e *Element) error {
	if e.handle == nil {
		return ErrInvalidElement
	}

	ref := e.handle.(C.AXUIElementRef)

	cName := C.CString("AXChildren")
	defer C.free(unsafe.Pointer(cName))

	attrName := C.cstring_to_cf_string(cName)
	defer C.CFRelease(C.CFTypeRef(attrName))

	value := C.ax_copy_attribute_value(ref, attrName)
	if value == 0 {
		e.Children = []*Element{}
		return nil
	}
	defer C.CFRelease(value)

	// Check if it's an array
	typeID := C.CFGetTypeID(value)
	if typeID != C.CFArrayGetTypeID() {
		e.Children = []*Element{}
		return nil
	}

	array := C.CFArrayRef(value)
	count := C.CFArrayGetCount(array)

	e.Children = make([]*Element, 0, int(count))

	// Create a temporary finder to use elementFromRef
	// TODO: This is a bit awkward - we should refactor to avoid this
	finder := &darwinFinder{}

	for i := C.CFIndex(0); i < count; i++ {
		childRef := C.AXUIElementRef(C.CFArrayGetValueAtIndex(array, i))
		if childRef != 0 {
			// Retain the child ref since we're taking ownership
			C.CFRetain(C.CFTypeRef(childRef))
			child, err := finder.elementFromRef(childRef, e)
			if err == nil && child != nil {
				e.Children = append(e.Children, child)
			}
		}
	}

	return nil
}

// mapRole converts macOS accessibility role to our Role type
func mapRole(axRole string) Role {
	switch axRole {
	case "AXWindow":
		return RoleWindow
	case "AXButton":
		return RoleButton
	case "AXTextField":
		return RoleTextField
	case "AXTextArea":
		return RoleTextArea
	case "AXStaticText":
		return RoleStaticText
	case "AXCheckBox":
		return RoleCheckbox
	case "AXRadioButton":
		return RoleRadioButton
	case "AXList":
		return RoleList
	case "AXRow", "AXOutlineRow":
		return RoleListItem
	case "AXMenu":
		return RoleMenu
	case "AXMenuItem":
		return RoleMenuItem
	case "AXMenuBar":
		return RoleMenuBar
	case "AXToolbar":
		return RoleToolbar
	case "AXScrollArea":
		return RoleScrollArea
	case "AXScrollBar":
		return RoleScrollBar
	case "AXImage":
		return RoleImage
	case "AXLink":
		return RoleLink
	case "AXGroup":
		return RoleGroup
	case "AXTabGroup":
		return RoleTabGroup
	case "AXTable":
		return RoleTable
	case "AXColumn":
		return RoleColumn
	case "AXCell":
		return RoleCell
	case "AXSlider":
		return RoleSlider
	case "AXComboBox":
		return RoleComboBox
	case "AXPopUpButton":
		return RolePopUpButton
	case "AXProgressIndicator":
		return RoleProgressBar
	case "AXSplitter":
		return RoleSplitter
	case "AXSheet":
		return RoleSheet
	case "AXDrawer":
		return RoleDrawer
	case "AXDialog":
		return RoleDialog
	case "AXApplication":
		return RoleApplication
	default:
		return RoleUnknown
	}
}
