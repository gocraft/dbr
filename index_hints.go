package dbr

//mysql hint https://dev.mysql.com/doc/refman/8.0/en/index-hints.html
const (
	useIndex    = "USE INDEX"
	ignoreIndex = "IGNORE INDEX"
	forceIndex  = "FORCE INDEX"

	forJoin    = "FOR JOIN"
	forOrderBY = "FOR ORDER BY"
	forGroupBy = "FOR GROUP BY"
)

type indexHint struct {
	hint      []string
	indexList []string
}

func UseIndex(index ...string) indexHint {
	return indexHint{hint: []string{useIndex}, indexList: index}
}

func IgnoreIndex(index ...string) indexHint {
	return indexHint{hint: []string{ignoreIndex}, indexList: index}
}

func ForceIndex(index ...string) indexHint {
	return indexHint{hint: []string{forceIndex}, indexList: index}
}

func (i indexHint) ForJoin() indexHint {
	i.hint = append(i.hint, forJoin)
	return i
}

func (i indexHint) ForOrderBy() indexHint {
	i.hint = append(i.hint, forOrderBY)
	return i
}
func (i indexHint) ForGroupBy() indexHint {
	i.hint = append(i.hint, forGroupBy)
	return i
}

func (i indexHint) Build(d Dialect, buf Buffer) error {
	for idx, v := range i.hint {
		if idx > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(v)
	}

	buf.WriteString("(")
	for idx, v := range i.indexList {
		if idx > 0 {
			buf.WriteString(",")
		}

		buf.WriteString(d.QuoteIdent(v))

	}
	buf.WriteString(")")
	return nil
}
