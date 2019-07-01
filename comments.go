package dbr

import "strings"

const (
	OPENING_SQL_COMMENT = "/*"
	CLOSING_SQL_COMMENT = "*/"
	SPACE               = " "
	NEWLINE             = "\n"
	EMPTY_STRING        = ""
)

// Comments represents a set of mysql comments
type Comments []string

func (comments Comments) Append(comment string) Comments {
	comment = strings.Replace(comment, OPENING_SQL_COMMENT, EMPTY_STRING, -1)
	comment = strings.Replace(comment, CLOSING_SQL_COMMENT, EMPTY_STRING, -1)
	comment = strings.TrimSpace(comment)
	comments = append(comments, comment)
	return comments
}

// Write will write each comment in the form of "/* some comment */\n"
func (comments Comments) Write(buf Buffer) {
	for _, comment := range comments {
		buf.WriteString(OPENING_SQL_COMMENT)
		buf.WriteString(SPACE)
		buf.WriteString(comment)
		buf.WriteString(SPACE)
		buf.WriteString(CLOSING_SQL_COMMENT)
		buf.WriteString(NEWLINE)
	}
}
