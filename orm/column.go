package orm

func (Column) expr() {}

func (Column) selectable() {}

type Column struct {
	name string
}

func C(name string) Column {
	return Column{
		name: name,
	}
}
