package dbr

import (
	"github.com/gocraft/dbr/v2/dialect"
)

// Dialect abstracts database driver differences in encoding
// types, and placeholders.
type Dialect = dialect.Dialect
