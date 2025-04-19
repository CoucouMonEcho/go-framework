package orm

type DB struct {
	r *registry
}

type DBOption func(db *DB)

func NewDB(opts ...DBOption) (*DB, error) {
	res := &DB{
		r: newRegistry(),
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func MustNewDB(opts ...DBOption) *DB {
	res, err := NewDB(opts...)
	if err != nil {
		panic(err)
	}
	return res
}
