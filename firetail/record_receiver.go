package firetail

import "log"

// recordReceiver receives records from the client into batches & passes them to the batch callback. If the batch callback
// returns an err, it does not remove the log entries from the batch.
func RecordReceiver(recordsChannel chan Record, maxBatchSize int, firetailApiUrl, firetailApiToken string) {
	recordsBatch := []Record{}

	for {
		newRecords, recordsRemaining := receiveRecords(recordsChannel, maxBatchSize-len(recordsBatch))
		recordsBatch = append(recordsBatch, newRecords...)

		// If the batch is empty, but there's records remaining, then we continue; else we return.
		if len(recordsBatch) == 0 {
			if recordsRemaining {
				continue
			} else {
				return
			}
		}

		// Give the batch to the batch callback. If it errs, we continue
		recordsSent, err := SendRecordsToSaaS(recordsBatch, firetailApiUrl, firetailApiToken)
		if err != nil {
			log.Println("Error sending records to Firetail:", err.Error())
			continue
		}
		log.Println("Successfully sent", recordsSent, "record(s) to Firetail.")

		// If the batch callback succeeded, we can clear the batch!
		recordsBatch = []Record{}
	}
}

// ReceiveRecords returns a slice of firetail Records up to the size of `limit`, and a boolean indicating that the channel
// still has items to be read - it will only be `false` when the channel is closed & empty. It achieves this by continuously
// reading from the log server's recordsChannel until it's empty, or the size limit has been reached.
func receiveRecords(recordsChannel chan Record, limit int) ([]Record, bool) {
	records := []Record{}
	for {
		select {
		case record, open := <-recordsChannel:
			if !open {
				return records, false
			}
			records = append(records, record)
			if len(records) == limit {
				return records, true
			}
		default:
			return records, true
		}
	}
}
