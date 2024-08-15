package main

import (
	"sync"

	"github.com/sirupsen/logrus"
)

func example() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.Info("EXAMPLE NO PARAMS")
	logrus.Info("READ FILE")
	var wga sync.WaitGroup

	cha := make(chan NewRow)
	fa := NewReadFascade("docs/моя.xlsx", "TDSheet")

	paramsa, err := fa.NewReadFull()
	if err != nil {
		panic(err)
	}

	wga.Add(1)
	go paramsa.ReadRows(cha, &wga)

	wga.Add(1)
	go func() {
		defer wga.Done()

		for row := range cha {
			logrus.Debug(row)
		}
	}()

	wga.Wait()
	logrus.Info("end")

	logrus.Info("READ FILE SYNC")
	resulta, err := paramsa.ReadRowsSync()
	if err != nil {
		logrus.Error(err)
	}

	logrus.Debug(resulta)

	////////////////////////////////////////////////////
	logrus.Info("EXAMPLE WITH PARAMS")
	logrus.Info("READ FILE")
	var wg sync.WaitGroup

	ch := make(chan NewRow)
	f := NewReadFascade("docs/моя.xlsx", "TDSheet")
	colParams := make([]ColumnParam, 0)
	colParams = append(colParams, ColumnParam{Id: 0, Required: true})
	//colParams = append(colParams, ColumnParam{Id: 1, Required: false})
	colParams = append(colParams, ColumnParam{Id: 2, Required: false})
	colParams = append(colParams, ColumnParam{Id: 3, Required: false})
	params := f.NewReadWithParams(colParams, 10)

	wg.Add(1)
	go params.ReadRows(ch, &wg)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for row := range ch {
			logrus.Debug(row)
		}
	}()

	wg.Wait()
	logrus.Info("end")

	logrus.Info("READ FILE SYNC")
	result, err := params.ReadRowsSync()
	if err != nil {
		logrus.Error(err)
	}

	logrus.Debug(result)
}
