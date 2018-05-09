package dbr

// ReturnInt64 executes the SelectStmt and returns the value as an int64.
func (b *SelectStmt) ReturnInt64() (int64, error) {
	var v int64
	err := b.LoadOne(&v)
	return v, err
}

// ReturnInt64s executes the SelectStmt and returns the value as a slice of int64s.
func (b *SelectStmt) ReturnInt64s() ([]int64, error) {
	var v []int64
	_, err := b.Load(&v)
	return v, err
}

// ReturnUint64 executes the SelectStmt and returns the value as an uint64.
func (b *SelectStmt) ReturnUint64() (uint64, error) {
	var v uint64
	err := b.LoadOne(&v)
	return v, err
}

// ReturnUint64s executes the SelectStmt and returns the value as a slice of uint64s.
func (b *SelectStmt) ReturnUint64s() ([]uint64, error) {
	var v []uint64
	_, err := b.Load(&v)
	return v, err
}

// ReturnString executes the SelectStmt and returns the value as a string.
func (b *SelectStmt) ReturnString() (string, error) {
	var v string
	err := b.LoadOne(&v)
	return v, err
}

// ReturnStrings executes the SelectStmt and returns the value as a slice of strings.
func (b *SelectStmt) ReturnStrings() ([]string, error) {
	var v []string
	_, err := b.Load(&v)
	return v, err
}
