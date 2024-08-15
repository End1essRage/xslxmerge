package xslxmerge

type ReadFascade struct {
	filePath  string
	sheetName string
}

func NewReadFascade(filePath string, sheetName string) *ReadFascade {
	return &ReadFascade{filePath: filePath, sheetName: sheetName}
}

func (f *ReadFascade) NewReadFull() (*ReadCommand, error) {
	cols, err := GetAllHeaders(f.filePath, f.sheetName)
	if err != nil {
		return nil, err
	}

	newParams := make([]ColumnParam, 0)
	for id := range cols {
		newParams = append(newParams, ColumnParam{Id: id})
	}

	return &ReadCommand{filePath: f.filePath, sheetName: f.sheetName, params: newParams, endRow: 0}, nil
}

func (f *ReadFascade) NewReadWithParams(params []ColumnParam, endRow int) *ReadCommand {
	return &ReadCommand{filePath: f.filePath, sheetName: f.sheetName, params: params, endRow: endRow}
}
