package dbr

type JoinType uint8

const (
	Inner JoinType = iota
	Left
	Right
	Full
)

func Join(t JoinType, table interface{}, on interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		buf.WriteString(" ")
		switch t {
		case Left:
			buf.WriteString("LEFT ")
		case Right:
			buf.WriteString("RIGHT ")
		case Full:
			buf.WriteString("FULL ")
		}
		buf.WriteString("JOIN ")
		switch table := table.(type) {
		case string:
			buf.WriteString(d.QuoteIdent(table))
		default:
			buf.WriteString(d.Placeholder())
			buf.WriteValue(table)
		}
		buf.WriteString(" ON ")
		switch on := on.(type) {
		case string:
			buf.WriteString(on)
		case Condition:
			buf.WriteString(d.Placeholder())
			buf.WriteValue(on)
		}
		return nil
	})
}
