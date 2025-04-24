package orm

type Column struct {
	name string
}

func (Column) expr() {}

func (Column) selectable() {}

func C(name string) Column {
	return Column{
		name: name,
	}
}
