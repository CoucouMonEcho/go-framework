package orm

type Deleter[T any] struct {
	builder
	table string

	db *DB

	where []Predicate
}

func NewDeleter[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		db: db,
	}
}

func (d *Deleter[T]) Build() (*Query, error) {
	var err error
	d.model, err = d.db.r.Register(new(T))
	if err != nil {
		return nil, err
	}

	d.sb.WriteString("DELETE FROM ")

	// table name
	if d.table == "" {
		d.sb.WriteByte('`')
		d.sb.WriteString(d.model.TableName)
		d.sb.WriteByte('`')
	} else {
		// d.sb.WriteByte('`')
		d.sb.WriteString(d.table)
		// d.sb.WriteByte('`')
	}

	// where condition
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		err = d.buildPredicates(d.where)
		if err != nil {
			return nil, err
		}
	}

	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deleter[T]) addArg(val any) *Deleter[T] {
	if d.args == nil {
		d.args = make([]any, 0, 4)
	}
	d.args = append(d.args, val)
	return d
}

// id := []int{1, 2, 3}
// wrong -> s.Where("id in (?, ?, ?)", ids)
// right -> s.Where("id in (?, ?, ?)", ids...)
func (d *Deleter[T]) Where(ps ...Predicate) *Deleter[T] {
	d.where = ps
	return d
}

func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}
