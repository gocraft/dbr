package dbr

//
// These are a set of helpers that just call LoadValue and return the value.
// They return (_, ErrNotFound) if nothing was found.
//

// The inclusion of these helpers in the package is not an obvious choice:
// Benefits:
//  - slight increase in code clarity/conciseness b/c you can use ":=" to define the variable
//
//    count, err := d.Select("COUNT(*)").From("users").Where("x = ?", x).ReturnInt64()
//
//    vs
//
//    var count int64
//    err := d.Select("COUNT(*)").From("users").Where("x = ?", x).LoadValue(&count)
//
// Downsides:
//  - very small increase in code cost, although it's not complex code
//  - increase in conceptual model / API footprint when presenting the package to new users
//  - no functionality that you can't achieve calling .LoadValue directly.
//  - There's a lot of possible types. Do we want to include ALL of them? u?int{8,16,32,64}?, strings, null varieties, etc.
//    - Let's just do the common, non-null varieties.

func (b *SelectBuilder) ReturnInt64() (int64, error) {
	var v int64
	err := b.LoadValue(&v)
	return v, err
}

func (b *SelectBuilder) ReturnString() (string, error) {
	var v string
	err := b.LoadValue(&v)
	return v, err
}
