package main

type ReadFascade struct {
	filePath  string
	sheetName string
}

func NewReadFascade(filePath string, sheetName string) *ReadFascade {
	return &ReadFascade{filePath: filePath, sheetName: sheetName}
}

func (f *ReadFascade) NewReadFull() *ReadCommand {
	colParams := make([]ColumnParam, 0)

	return &ReadCommand{FilePath: f.filePath, SheetName: f.sheetName, Params: colParams, EndRow: 0}
}

func (f *ReadFascade) NewReadWithParams(params []ColumnParam, endRow int) *ReadCommand {
	return &ReadCommand{FilePath: f.filePath, SheetName: f.sheetName, Params: params, EndRow: endRow}
}
