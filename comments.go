package dbr

import "strings"

const (
	openingSQLComment = "/*"
	closingSQLComment = "*/"
	space             = " "
	newline           = "\n"
	emptyString       = ""
)

// Comments represents a set of sql comments
type Comments []string

// Append a new sql comment to a set of comments
func (comments Comments) Append(comment string) Comments {
	comment = strings.Replace(comment, openingSQLComment, emptyString, -1)
	comment = strings.Replace(comment, closingSQLComment, emptyString, -1)
	comment = strings.TrimSpace(comment)
	comments = append(comments, comment)
	return comments
}

// Build writes each comment in the form of "/* some comment */\n"
func (comments Comments) Build(d Dialect, buf Buffer) error {
	for _, comment := range comments {
		words := []string{openingSQLComment, space, comment, space, closingSQLComment, newline}
		for _, str := range words {
			if _, err := buf.WriteString(str); err != nil {
				return err
			}
		}
	}
	return nil
}
