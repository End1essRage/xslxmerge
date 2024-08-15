package main

import "github.com/sirupsen/logrus"

type ReadFascade struct {
	filePath  string
	sheetName string
}

func NewReadFascade(filePath string, sheetName string) *ReadFascade {
	return &ReadFascade{filePath: filePath, sheetName: sheetName}
}

func (f *ReadFascade) NewReadFull() *ReadCommand {
	cols, err := GetAllHeaders(f.filePath, f.sheetName)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Info(cols)
	newParams := make([]ColumnParam, 0)
	for id := range cols {
		newParams = append(newParams, ColumnParam{Id: id})
	}
	logrus.Info(newParams)

	return &ReadCommand{FilePath: f.filePath, SheetName: f.sheetName, Params: newParams, EndRow: 0}
}

func (f *ReadFascade) NewReadWithParams(params []ColumnParam, endRow int) *ReadCommand {
	return &ReadCommand{FilePath: f.filePath, SheetName: f.sheetName, Params: params, EndRow: endRow}
}
