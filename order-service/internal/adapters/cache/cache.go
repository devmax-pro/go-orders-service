package cache

type Value any
type Key any

type Cache[T Value] interface {
	Set(key Key, value T) error
	Get(key Key) (T, bool)
}
