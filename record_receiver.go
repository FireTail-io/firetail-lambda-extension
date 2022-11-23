package main

import (
	"firetail-lambda-extension/firetail"
	"firetail-lambda-extension/logsapiclient"
	"log"
	"sync"
)

// recordReceiver receives records from the logServer into batches & attempts to send them to Firetail until
// logServer.ReceiveRecords returns that there are no records remaining.
func recordReceiver(logServer *logsapiclient.Client, wg *sync.WaitGroup) {
	firetailApiToken, firetailApiUrl := getFiretailApiConfig()

	recordsBatch := []firetail.Record{}
	maxBatchSize := 100

	for {
		// Spend no more than 1 second receiving records into the recordsBatch
		newRecords, recordsRemaining := logServer.ReceiveRecords(maxBatchSize - len(recordsBatch))
		recordsBatch = append(recordsBatch, newRecords...)

		// If the batch is empty, but there's records remaining, then we continue; else we return.
		if len(recordsBatch) == 0 {
			if recordsRemaining {
				continue
			} else {
				wg.Done()
				return
			}
		}

		// Try to send the batch to Firetail
		log.Printf("Attempting to send batch of %d record(s) to Firetail...", len(recordsBatch))
		recordsSent, err := firetail.SendRecordsToSaaS(recordsBatch, firetailApiUrl, firetailApiToken)
		log.Printf("%d record(s) sent to Firetail.", recordsSent)
		if err != nil {
			log.Printf("Err sending record(s) to Firetail SaaS, err: %s", err.Error())
			continue
		}

		// If sending the batch to Firetail was a success, we can clear out the batch!
		log.Printf("Clearing records batch...")
		recordsBatch = []firetail.Record{}
	}
}
