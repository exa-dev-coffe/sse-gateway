package enum

// ParseGeneric → untuk convert string ke enum via map
func ParseGeneric[T ~string](s string, mapping map[string]T) (T, bool) {
	val, ok := mapping[s]
	return val, ok
}
