package dbr

var NameMapping = camelCaseToSnakeCase

func camelCaseToSnakeCase(name string) string {
  newstr := make([]rune, 0)
  firstTime := true

  for _, chr := range name {
    if isUpper := 'A' <= chr && chr <= 'Z'; isUpper {
      if firstTime == true {
        firstTime = false
      } else {
        newstr = append(newstr, '_')
      }
      chr -= ('A' - 'a')
    }
    newstr = append(newstr, chr)
  }

  return string(newstr)
}