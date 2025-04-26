package orm

type Assignment struct {
	col string
	val any
}

func (Assignment) assign() {}

func Assign(col string, val any) Assignment {
	return Assignment{col, val}
}
