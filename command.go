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

func (params ReadCommand) ReadRows(ch chan<- Row, wg *sync.WaitGroup) {
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
		rows, err := f.GetRows(params.SheetName)
		if err != nil {
			logrus.Error(err)
			return
		}

		for id, row := range rows[1:] {
			ch <- Row{id + 2: row}
		}

	} else {
		filledCells := params.fillCellsMap(f)
		params.sendRow(ch, filledCells)
	}
}

func (params ReadCommand) ReadRowsSync() ([]Row, error) {
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
		rows, err := f.GetRows(params.SheetName)
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		for id, row := range rows[1:] {
			result = append(result, Row{id + 2: row})
		}

	} else {
		filledCells := params.fillCellsMap(f)
		for i := 2; i < params.EndRow; i++ {
			row, err := params.getRow(i, filledCells)
			if err != nil {
				continue
			}
			result = append(result, Row{i: row})
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, fmt.Errorf("no rows")
}

func (params ReadCommand) fillCellsMap(file *excelize.File) map[string]string {
	filledCells := make(map[string]string)

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
				//logrus.Error("empty value")
			}

			// Заполняем мапу значением (пустая ячейка сохраняется как пустая строка)
			filledCells[cellRef] = cellValue
		}
	}

	return filledCells
}

func (params ReadCommand) sendRow(ch chan<- Row, filledCells map[string]string) {
	for i := 2; i < params.EndRow; i++ {
		row := make([]string, 0)

		skip := false
		for j := 0; j < len(params.Params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+params.Params[j].Id), i)

			if params.Params[j].Required && filledCells[cellRef] == emptyValue {
				skip = true
			}

			row = append(row, filledCells[cellRef])
		}

		if !skip {
			ch <- Row{i: row}
		}
	}
}

func (params ReadCommand) getRow(i int, filledCells map[string]string) ([]string, error) {
	row := make([]string, 0)

	skip := false
	for j := 0; j < len(params.Params); j++ {
		cellRef := fmt.Sprintf("%s%d", string('A'+params.Params[j].Id), i)

		if params.Params[j].Required && filledCells[cellRef] == emptyValue {
			skip = true
		}

		row = append(row, filledCells[cellRef])
	}

	if !skip {
		return row, nil
	}

	return nil, fmt.Errorf("row skipped")
}
