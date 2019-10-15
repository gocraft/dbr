package dbr

type union struct {
	builder []Builder
	all     bool
}

// Union builds `... UNION ...`.
func Union(builder ...Builder) interface {
	Builder
	As(string) Builder
} {
	return &union{
		builder: builder,
	}
}

// UnionAll builds `... UNION ALL ...`.
func UnionAll(builder ...Builder) interface {
	Builder
	As(string) Builder
} {
	return &union{
		builder: builder,
		all:     true,
	}
}

func (u *union) Build(d Dialect, buf Buffer) error {
	for i, b := range u.builder {
		if i > 0 {
			buf.WriteString(" UNION ")
			if u.all {
				buf.WriteString("ALL ")
			}
		}
		err := b.Build(d, buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *union) As(alias string) Builder {
	return as(u, alias)
}
