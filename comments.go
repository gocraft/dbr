package dbr

import "strings"

const (
	openingSQLComment = "/*"
	closingSQLComment = "*/"
	space             = " "
	newline           = "\n"
	emptyString       = ""
)

// Comments represents a set of mysql comments
type Comments []string

// Append a new sql comment to a set of comments
func (comments Comments) Append(comment string) Comments {
	comment = strings.Replace(comment, openingSQLComment, emptyString, -1)
	comment = strings.Replace(comment, closingSQLComment, emptyString, -1)
	comment = strings.TrimSpace(comment)
	comments = append(comments, comment)
	return comments
}

// Write will write each comment in the form of "/* some comment */\n"
func (comments Comments) Write(buf Buffer) {
	for _, comment := range comments {
		buf.WriteString(openingSQLComment)
		buf.WriteString(space)
		buf.WriteString(comment)
		buf.WriteString(space)
		buf.WriteString(closingSQLComment)
		buf.WriteString(newline)
	}
}
