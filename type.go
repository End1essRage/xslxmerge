package xslxmerge

type ColumnParam struct {
	Id       int
	Required bool
}

type Headers []Cell

type Cell struct {
	ColumnId int
	Data     string
}

type Row struct {
	Id    int
	Empty bool
	Data  []Cell
}
