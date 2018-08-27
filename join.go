package dbr

type joinType uint8

const (
	inner joinType = iota
	allFull
	left
	right
	full
	anyLeft
)

func join(t joinType, table interface{}, on interface{}) Builder {
	return BuildFunc(func(d Dialect, buf Buffer) error {
		buf.WriteString(" ")
		switch t {
		case anyLeft:
			buf.WriteString("ANY LEFT ")
		case allFull:
			buf.WriteString("ALL FULL ")
		case left:
			buf.WriteString("LEFT ")
		case right:
			buf.WriteString("RIGHT ")
		case full:
			buf.WriteString("FULL ")
		}
		buf.WriteString("JOIN ")
		switch table := table.(type) {
		case string:
			buf.WriteString(d.QuoteIdent(table))
		default:
			buf.WriteString(placeholder)
			buf.WriteValue(table)
		}
		if d.SupportsOn() {
			buf.WriteString(" ON ")
		} else {
			buf.WriteString(" USING ")
		}
		switch on := on.(type) {
		case string:
			buf.WriteString(on)
		case Builder:
			buf.WriteString(placeholder)
			buf.WriteValue(on)
		}
		return nil
	})
}
