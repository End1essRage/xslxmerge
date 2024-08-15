package main

type ColumnParam struct {
	Id       int
	Required bool
	//format
}

// id столбца + имя
type Columns map[int]string

// номер строки + набор данных
type Row map[int]interface{}
