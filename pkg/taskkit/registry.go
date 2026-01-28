package taskkit

import (
	"fmt"
	"sort"
	"sync"
)

// StepHandler is the function signature for step implementations
type StepHandler func(input StepInput, deps Deps) StepResult

var (
	registry     = make(map[string]StepHandler)
	registryLock sync.RWMutex
)

// Register adds a step handler to the global registry.
// This is typically called from init() functions in step packages.
// Panics if a handler with the same name is already registered.
func Register(name string, handler StepHandler) {
	registryLock.Lock()
	defer registryLock.Unlock()

	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("step handler already registered: %s", name))
	}
	registry[name] = handler
}

// Get retrieves a step handler by name
func Get(name string) (StepHandler, bool) {
	registryLock.RLock()
	defer registryLock.RUnlock()

	handler, ok := registry[name]
	return handler, ok
}

// MustGet retrieves a step handler by name, panicking if not found
func MustGet(name string) StepHandler {
	handler, ok := Get(name)
	if !ok {
		panic(fmt.Sprintf("step handler not found: %s", name))
	}
	return handler
}

// ListHandlers returns all registered handler names
func ListHandlers() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HandlerCount returns the number of registered handlers
func HandlerCount() int {
	registryLock.RLock()
	defer registryLock.RUnlock()

	return len(registry)
}
