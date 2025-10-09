package regularintmap

import "sync"

// RegularIntMap is a map[int]int protected by an RWMutex.
type RegularIntMap struct {
    sync.RWMutex
    internal map[int]int
}

// NewRegularIntMap constructs a RegularIntMap with an initialized map.
func NewRegularIntMap() *RegularIntMap {
    return &RegularIntMap{
        internal: make(map[int]int),
    }
}

// Load returns the value for a key and whether it was present.
func (rm *RegularIntMap) Load(key int) (value int, ok bool) {
    rm.RLock()
    result, ok := rm.internal[key]
    rm.RUnlock()
    return result, ok
}

// Delete removes the key from the map.
func (rm *RegularIntMap) Delete(key int) {
    rm.Lock()
    delete(rm.internal, key)
    rm.Unlock()
}

// Store sets the value for a key.
func (rm *RegularIntMap) Store(key, value int) {
    rm.Lock()
    rm.internal[key] = value
    rm.Unlock()
}
