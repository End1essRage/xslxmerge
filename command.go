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
const checkFirstStep = 10000

func (c *ReadCommand) Read() ReadResult {
	result := ReadResult{}

	ids := make([]int, 0)
	for _, val := range c.Params {
		ids = append(ids, val.Id)
	}

	headers, err := GetHeaders(c.FilePath, c.SheetName, ids)
	if err != nil {
		result.Error = err
		return result
	}

	result.Headers = headers

	rows, err := c.ReadRowsSync()
	if err != nil {
		result.Error = err
		return result
	}

	result.Rows = rows

	return result
}

func (c *ReadCommand) ReadRows(ch chan<- Row, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(ch)

	f, err := excelize.OpenFile(c.FilePath)
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

	filledCells := c.fillCellsMap(f)
	c.sendRow(ch, filledCells)
}

func (c *ReadCommand) ReadRowsSync() ([]Row, error) {
	f, err := excelize.OpenFile(c.FilePath)
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

	filledCells := c.fillCellsMap(f)
	for i := startRow; i < c.EndRow; i++ {
		row := c.getRow(i, filledCells)
		if row.Empty {
			continue
		}

		result = append(result, row)
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, fmt.Errorf("no rows")
}

func (c *ReadCommand) findEndRow(file *excelize.File) int {
	empty := false
	emptyCounter := 0
	currentStep := checkFirstStep
	counter := currentStep
	lastEmpty := 0

	for {
		emptyCounter = 0
		// TODO Необходимо переделать на если хотя бы в одной ячейке есть значение остальное пропускать
		for j := 0; j < len(c.Params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+c.Params[j].Id), counter) // Определяем ссылку на ячейку

			// Проверяем, есть ли значение в строке или нет
			cellValue, err := file.GetCellValue(c.SheetName, cellRef)
			if err != nil {
				cellValue = emptyValue
			}

			if cellValue != "" {
				break
			} else {
				emptyCounter++
				cellValue = emptyValue
			}

			if emptyCounter == len(c.Params) {
				empty = true
			}
		}

		if empty {
			if currentStep >= 4 {
				currentStep /= 4
			}
			lastEmpty = counter
			counter -= currentStep * 2
		} else {
			counter += currentStep
		}

		logrus.Info("Поиск последней строки : ")
		logrus.Infof("last empty = %i \n counter = %i \n check step = %i", lastEmpty, counter, currentStep)

		if counter == lastEmpty-1 {
			break
		}
		/*if end {
			break
		}
		counter++
		*/
	}

	return counter
}

func (c *ReadCommand) fillCellsMap(file *excelize.File) map[string]string {
	filledCells := make(map[string]string)

	if c.EndRow == 0 {
		c.EndRow = c.findEndRow(file)
	}

	for i := startRow; i < c.EndRow; i++ {
		for j := 0; j < len(c.Params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+c.Params[j].Id), i) // Определяем ссылку на ячейку

			// Проверяем, есть ли значение в строке или нет
			cellValue, err := file.GetCellValue(c.SheetName, cellRef)
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

func (c *ReadCommand) sendRow(ch chan<- Row, filledCells map[string]string) {
	for i := startRow; i < c.EndRow; i++ {
		row := Row{Id: i}

		skip := false
		for j := 0; j < len(c.Params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+c.Params[j].Id), i)

			if c.Params[j].Required && filledCells[cellRef] == emptyValue {
				skip = true
			}
			cell := Cell{ColumnId: c.Params[j].Id, Data: filledCells[cellRef]}
			row.Data = append(row.Data, cell)
		}

		if !skip {
			ch <- row
		}
	}
}

func (c *ReadCommand) getRow(i int, filledCells map[string]string) Row {
	row := Row{Id: i}
	skip := false
	for j := 0; j < len(c.Params); j++ {
		cellRef := fmt.Sprintf("%s%d", string('A'+c.Params[j].Id), i)

		if c.Params[j].Required && filledCells[cellRef] == emptyValue {
			skip = true
		}
		cell := Cell{ColumnId: c.Params[j].Id, Data: filledCells[cellRef]}
		row.Data = append(row.Data, cell)
	}

	if !skip {
		return row
	}
	row.Empty = true

	return row
}
