package main

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type ReadCommand struct {
	SheetName string
	FilePath  string
	Params    []ColumnParam
	EndRow    int
}

const emptyValue = "EMPTY"
const startRow = 2

func (params *ReadCommand) ReadRows(ch chan<- Rows, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(ch)

	f, err := excelize.OpenFile(params.FilePath)
	if err != nil {
		logrus.Error(err)
		return
	}

	defer func() {
		// Закрываем таблицу
		if err := f.Close(); err != nil {
			logrus.Error(err)
		}
	}()

	filledCells := params.fillCellsMap(f)
	params.sendRow(ch, filledCells)
}

func (params *ReadCommand) ReadRowsSync() ([]Rows, error) {
	f, err := excelize.OpenFile(params.FilePath)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	defer func() {
		// Закрываем таблицу
		if err := f.Close(); err != nil {
			logrus.Error(err)
		}
	}()

	result := make([]Rows, 0)

	filledCells := params.fillCellsMap(f)
	for i := startRow; i < params.EndRow; i++ {
		row, err := params.getRow(i, filledCells)
		if err != nil {
			continue
		}
		result = append(result, Rows{i: row})
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, fmt.Errorf("no rows")
}

func (params *ReadCommand) findEndRow(file *excelize.File) int {
	counter := startRow
	end := false
	emptyCounter := 0
	for {
		emptyCounter = 0
		// TODO Необходимо переделать на если хотя бы в одной ячейке есть значение остальное пропускать
		for j := 0; j < len(params.Params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+params.Params[j].Id), counter) // Определяем ссылку на ячейку

			// Проверяем, есть ли значение в строке или нет
			cellValue, err := file.GetCellValue(params.SheetName, cellRef)
			if err != nil {
				cellValue = emptyValue
			}

			if cellValue != "" {
				break
			} else {
				emptyCounter++
				cellValue = emptyValue
			}

			if emptyCounter == len(params.Params) {
				end = true
			}
		}
		if end {
			break
		}
		counter++
	}

	return counter
}

func (params *ReadCommand) fillCellsMap(file *excelize.File) map[string]string {
	filledCells := make(map[string]string)

	if params.EndRow == 0 {
		params.EndRow = params.findEndRow(file)
	}

	for i := startRow; i < params.EndRow; i++ {
		for j := 0; j < len(params.Params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+params.Params[j].Id), i) // Определяем ссылку на ячейку

			// Проверяем, есть ли значение в строке или нет
			cellValue, err := file.GetCellValue(params.SheetName, cellRef)
			if err != nil {
				cellValue = emptyValue
			}

			if cellValue == "" {
				cellValue = emptyValue
			}

			// Заполняем мапу значением (пустая ячейка сохраняется как пустая строка)
			filledCells[cellRef] = cellValue
		}
	}

	return filledCells
}

func (params *ReadCommand) sendRow(ch chan<- Rows, filledCells map[string]string) {
	for i := startRow; i < params.EndRow; i++ {
		row := make(RowData)

		skip := false
		for j := 0; j < len(params.Params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+params.Params[j].Id), i)

			if params.Params[j].Required && filledCells[cellRef] == emptyValue {
				skip = true
			}

			row[params.Params[j].Id] = filledCells[cellRef]
		}

		if !skip {
			ch <- Rows{i: row}
		}
	}
}

func (params *ReadCommand) getRow(i int, filledCells map[string]string) (RowData, error) {
	row := make(RowData)

	skip := false
	for j := 0; j < len(params.Params); j++ {
		cellRef := fmt.Sprintf("%s%d", string('A'+params.Params[j].Id), i)

		if params.Params[j].Required && filledCells[cellRef] == emptyValue {
			skip = true
		}

		row[params.Params[j].Id] = filledCells[cellRef]
	}

	if !skip {
		return row, nil
	}

	return nil, fmt.Errorf("row skipped")
}
