package main

import "github.com/influxdb/influxdb-go"

func NewInfluxCollector(dbCfg *influxdb.ClientConfig) (chan<- Serieses, error) {
	var (
		err error
		db  *influxdb.Client
	)

	//dbCfg.Database = "usage"

	db, err = influxdb.NewClient(dbCfg)
	if err != nil {
		return nil, err
	}

	seriesChan := make(chan Serieses)

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
