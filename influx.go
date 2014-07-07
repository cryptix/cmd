package main

import "github.com/influxdb/influxdb-go"

type serieses []*influxdb.Series

func NewInfluxCollector(dbCfg *influxdb.ClientConfig) (chan<- serieses, error) {
	var (
		err error
		db  *influxdb.Client
	)

	//dbCfg.Database = "usage"

	db, err = influxdb.NewClient(dbCfg)
	if err != nil {
		return nil, err
	}

	seriesChan := make(chan serieses)

	// start a goroutine that sends to influxdb
	go func() {
		var err error
		// loop over
		for series := range seriesChan {

			err = db.WriteSeries(series)
			checkFatal(err)

		}
	}()

	return seriesChan, nil
}
