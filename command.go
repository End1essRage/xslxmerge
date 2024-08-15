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

func (params *ReadCommand) ReadRows(ch chan<- Row, wg *sync.WaitGroup) {
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

	if len(params.Params) < 1 {
		//Переделать на заполнение params по колву столбоц в хедере и реализовать вывод с первым столбцом
		cols, err := GetAllHeaders(params.FilePath, params.SheetName)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Info(cols)
		newParams := make([]ColumnParam, 0)
		for id := range cols {
			newParams = append(newParams, ColumnParam{Id: id})
		}
		logrus.Info(newParams)
		params.Params = newParams
	}
	logrus.Info(len(params.Params))
	filledCells := params.fillCellsMap(f)
	params.sendRow(ch, filledCells)
}

func (params *ReadCommand) ReadRowsSync() ([]Row, error) {
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

	result := make([]Row, 0)

	if len(params.Params) < 1 {
		cols, err := GetAllHeaders(params.FilePath, params.SheetName)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Info(cols)
		newParams := make([]ColumnParam, 0)
		for id := range cols {
			newParams = append(newParams, ColumnParam{Id: id})
		}
		logrus.Info(newParams)
		params.Params = newParams

	}
	filledCells := params.fillCellsMap(f)
	for i := 2; i < params.EndRow; i++ {
		row, err := params.getRow(i, filledCells)
		if err != nil {
			continue
		}
		result = append(result, Row{i: row})
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, fmt.Errorf("no rows")
}

func (params *ReadCommand) fillCellsMap(file *excelize.File) map[string]string {
	filledCells := make(map[string]string)
	// find last row
	if params.EndRow == 0 {
		counter := 2
		end := false
		emptyCounter := 0
		for {
			emptyCounter = 0
			for j := 0; j < len(params.Params); j++ {
				cellRef := fmt.Sprintf("%s%d", string('A'+params.Params[j].Id), counter) // Определяем ссылку на ячейку

				// Проверяем, есть ли значение в строке или нет
				cellValue, err := file.GetCellValue(params.SheetName, cellRef)
				if err != nil {
					cellValue = emptyValue
				}

				if cellValue == "" {
					cellValue = emptyValue
				}

				if cellValue == emptyValue {
					emptyCounter++
				}

				if emptyCounter == len(params.Params) {
					end = true
				}
				// Заполняем мапу значением (пустая ячейка сохраняется как пустая строка)
				filledCells[cellRef] = cellValue
			}
			if end {
				break
			}
			counter++
		}

		params.EndRow = counter
	} else {
		for i := 2; i < params.EndRow; i++ {
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
	}

	return filledCells
}

func (params *ReadCommand) sendRow(ch chan<- Row, filledCells map[string]string) {
	for i := 2; i < params.EndRow; i++ {
		row := make(map[int]string)

		skip := false
		for j := 0; j < len(params.Params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+params.Params[j].Id), i)

			if params.Params[j].Required && filledCells[cellRef] == emptyValue {
				skip = true
			}

			row[params.Params[j].Id] = filledCells[cellRef]
		}

		if !skip {
			ch <- Row{i: row}
		}
	}
}

func (params *ReadCommand) getRow(i int, filledCells map[string]string) (map[int]string, error) {
	row := make(map[int]string)

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
