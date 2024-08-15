package main

type ColumnParam struct {
	Id       int
	Required bool
}

// id столбца + имя
type Headers map[int]string

// номер строки + набор данных
type Rows map[int]RowData

// номер столбца + данные ячейки
type RowData map[int]string
