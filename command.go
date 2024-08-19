package main

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type ReadCommand struct {
	sheetName string
	filePath  string
	params    []ColumnParam
	endRow    int
}

const emptyValue = "EMPTY"
const startRow = 2
const checkFirstStep = 100000

func (c *ReadCommand) Read() ReadFullResult {
	result := ReadFullResult{}

	ids := make([]int, 0)
	for _, val := range c.params {
		ids = append(ids, val.Id)
	}

	headers, err := GetHeaders(c.filePath, c.sheetName, ids)
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

	f, err := excelize.OpenFile(c.filePath)
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
	f, err := excelize.OpenFile(c.filePath)
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
	for i := startRow; i < c.endRow; i++ {
		row := c.getRow(i, filledCells)
		if row.Empty {
			continue
		}

		result = append(result, row)
	}

	if len(result) > 0 {
		return result, nil
	}

	return result, nil
}

func (c *ReadCommand) findEndRow(file *excelize.File) int {
	empty := false
	emptyCounter := 0
	currentStep := checkFirstStep
	counter := currentStep
	topFilled := 0
	bottomEmpty := 0
	params := len(c.params)
	iterationsCount := 0

	for {
		iterationsCount += 1
		emptyCounter = 0

		// TODO Необходимо переделать на если хотя бы в одной ячейке есть значение остальное пропускать
		for j := 0; j < params; j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+c.params[j].Id), counter) // Определяем ссылку на ячейку

			// Проверяем, есть ли значение в строке или нет
			cellValue, err := file.GetCellValue(c.sheetName, cellRef)
			if err != nil {
				cellValue = emptyValue
			}

			if cellValue != "" {
				empty = false
				break
			} else {
				emptyCounter++
				cellValue = emptyValue
			}

			empty = emptyCounter == len(c.params)
		}

		if empty {
			currentStep /= 2

			if currentStep <= 2 {
				currentStep = 1
			}

			bottomEmpty = counter
			counter -= currentStep

		} else {
			topFilled = counter

			counter += currentStep
		}

		inRange := bottomEmpty - topFilled

		if inRange <= 1 {
			break
		}
	}

	return topFilled
}

func (c *ReadCommand) fillCellsMap(file *excelize.File) map[string]string {
	filledCells := make(map[string]string)

	if c.endRow == 0 {
		c.endRow = c.findEndRow(file)
	}

	for i := startRow; i < c.endRow; i++ {
		for j := 0; j < len(c.params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+c.params[j].Id), i) // Определяем ссылку на ячейку

			// Проверяем, есть ли значение в строке или нет
			cellValue, err := file.GetCellValue(c.sheetName, cellRef)
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
	for i := startRow; i < c.endRow; i++ {
		row := Row{Id: i}

		skip := false
		for j := 0; j < len(c.params); j++ {
			cellRef := fmt.Sprintf("%s%d", string('A'+c.params[j].Id), i)

			if c.params[j].Required && filledCells[cellRef] == emptyValue {
				skip = true
			}
			cell := Cell{ColumnId: c.params[j].Id, Data: filledCells[cellRef]}
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
	for j := 0; j < len(c.params); j++ {
		cellRef := fmt.Sprintf("%s%d", string('A'+c.params[j].Id), i)

		if c.params[j].Required && filledCells[cellRef] == emptyValue {
			skip = true
		}
		cell := Cell{ColumnId: c.params[j].Id, Data: filledCells[cellRef]}
		row.Data = append(row.Data, cell)
	}

	if !skip {
		return row
	}
	row.Empty = true

	return row
}
