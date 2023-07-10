package collections

import (
	"strings"
	"time"
)

type Fields map[string]any

func (f Fields) Get(key string) (any, bool) {
	rootKey, subKeys, split := strings.Cut(key, ".")
	if split {
		switch v := f[rootKey].(type) {
		case Fields:
			return v.Get(subKeys)
		case map[string]any:
			return Fields(v).Get(subKeys)
		default:
			rootKey = key
		}
	}

	value, ok := f[rootKey]

	if m, isMap := value.(map[string]any); ok && isMap {
		return Fields(m), ok
	}

	return value, ok
}

func (f Fields) GetString(key string) (string, bool) {
	value, ok := f.Get(key)
	if !ok {
		return "", false
	}
	str, ok := value.(string)
	if !ok {
		return "", false
	}

	return str, true
}

func (f Fields) GetNumber(key string) (float64, bool) {
	value, ok := f.Get(key)
	if !ok {
		return 0, false
	}
	num, ok := value.(float64)
	if !ok {
		return 0, false
	}

	return num, true
}

func (f Fields) GetBool(key string) (bool, bool) {
	value, ok := f.Get(key)
	if !ok {
		return false, false
	}

	b, ok := value.(bool)
	if !ok {
		return false, false
	}

	return b, true
}

func (f Fields) GetTime(key string, format string) (time.Time, bool) {
	value, ok := f.GetString(key)
	if !ok {
		return time.Time{}, false
	}

	t, err := time.Parse(format, value)
	if err != nil {
		return time.Time{}, false
	}

	return t, true
}

func (f Fields) Set(key string, value any) {
	rootKey, subKeys, split := strings.Cut(key, ".")
	if split {
		subMap, ok := f[rootKey].(Fields)
		if !ok {
			subMap = Fields{}
			subMap.Set(subKeys, value)
			f[rootKey] = subMap
		} else {
			subMap.Set(subKeys, value)
		}
	} else {
		f[rootKey] = value
	}
}

// Delete will remove the key from the Fields. If key is nested,
// empty sub-keys will be removed as well.
func (f Fields) Delete(key string) {
	rootKey, subKeys, split := strings.Cut(key, ".")
	if split {
		subMap, ok := f[rootKey].(Fields)
		if ok {
			subMap.Delete(subKeys)
			if len(subMap) == 0 {
				delete(f, rootKey)
			}
		}
	} else {
		delete(f, rootKey)
	}
}

func (f Fields) Walk(walkFn WalkFunc) {
	f.walk("", walkFn)
}

func (f Fields) walk(parentKey string, walkFn WalkFunc) {
	for k, v := range f {
		if subMap, ok := f[k].(Fields); ok {
			subMap.walk(k, walkFn)
		}
		fullKey := k
		if parentKey != "" {
			fullKey = parentKey + "." + fullKey
		}
		walkFn(f, fullKey, k, v)
	}
}

// WalkFunc is a callback function that is called for each value of Fields and its
// sub-nodes. The direct Fields that is associated with key and value is provided. The
// full path of the key is provided by fullKey.
type WalkFunc = func(evt Fields, fullKey, key string, value any)
