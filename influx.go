package main

import "github.com/influxdb/influxdb/client"

type serieses []*client.Series

func NewInfluxCollector(dbCfg *client.ClientConfig) (chan<- serieses, error) {
	var (
		err error
		db  *client.Client
	)

	//dbCfg.Database = "usage"

	db, err = client.NewClient(dbCfg)
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
