package tools

import "github.com/anxuanzi/cua/pkg/input"

// typeTextNative types text using robotgo.
func typeTextNative(text string) error {
	return input.TypeText(text)
}

// keyPressNative presses a key with optional modifiers using robotgo.
func keyPressNative(key string, modifiers []string) error {
	return input.KeyTapWithModifiers(key, modifiers)
}
