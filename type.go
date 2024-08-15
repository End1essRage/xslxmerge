package main

type ColumnParam struct {
	Id       int
	Required bool
}

// id столбца + имя
type Headers map[int]string
type NewHeaders []NewCell

/*
// номер строки + набор данных
type Rows map[int]RowData

// номер столбца + данные ячейки
type RowData map[int]string
*/
type NewRows []NewRow

type NewCell struct {
	ColumnId int
	Data     string
}

type NewRow struct {
	Id    int
	Empty bool
	Data  []NewCell
}
