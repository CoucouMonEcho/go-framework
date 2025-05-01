package orm

type Column struct {
	table TableReference
	name  string
	alias string
}

func (Column) expr() {}

func (Column) selectable() {}

func (Column) assign() {}

func C(name string) Column {
	return Column{
		name: name,
	}
}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
		table: c.table,
	}
}
