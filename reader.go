package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

func GetAllHeaders(filePath string, sheetName string) (Headers, error) {
	f, err := excelize.OpenFile(filePath)
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

	result := make(Headers, 0)
	counter := 0
	for {
		cellRef := fmt.Sprintf("%s%d", string('A'+counter), 1)
		cellValue, err := f.GetCellValue(sheetName, cellRef)
		if err != nil {
			break
		}

		if cellValue == "" {
			break
		}
		cell := Cell{ColumnId: counter, Data: cellValue}
		result = append(result, cell)

		counter++
	}

	return result, nil
}

func GetHeaders(filePath string, sheetName string, colIds []int) (Headers, error) {
	f, err := excelize.OpenFile(filePath)
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

	result := make(Headers, 0)

	for _, col := range colIds {
		cellRef := fmt.Sprintf("%s%d", string('A'+col), 1)
		cellValue, err := f.GetCellValue(sheetName, cellRef)
		if err != nil {
			cellValue = emptyValue
		}

		if cellValue == "" {
			cellValue = emptyValue
		}

		cell := Cell{ColumnId: col, Data: cellValue}
		result = append(result, cell)
	}

	return result, nil
}
