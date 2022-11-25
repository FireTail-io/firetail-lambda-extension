package logsapi

import (
	"firetail-lambda-extension/firetail"
)

// recordReceiver receives records from the client into batches & passes them to the batch callback. If the batch callback
// returns an err, it does not remove the log entries from the batch.
func (c *Client) recordReceiver() {
	recordsBatch := []firetail.Record{}

	for {
		newRecords, recordsRemaining := c.receiveRecords(c.maxBatchSize - len(recordsBatch))
		recordsBatch = append(recordsBatch, newRecords...)

		// If the batch is empty, but there's records remaining, then we continue; else we return.
		if len(recordsBatch) == 0 {
			if recordsRemaining {
				continue
			} else {
				c.receiverWaitgroup.Done()
				return
			}
		}

		// Give the batch to the batch callback. If it errs, we continue
		err := c.batchCallback(recordsBatch)
		if err != nil {
			c.errCallback(err)
			continue
		}

		// If the batch callback succeeded, we can clear the batch!
		recordsBatch = []firetail.Record{}
	}
}

// ReceiveRecords returns a slice of firetail Records up to the size of `limit`, and a boolean indicating that the channel
// still has items to be read - it will only be `false` when the channel is closed & empty. It achieves this by continuously
// reading from the log server's recordsChannel until it's empty, or the size limit has been reached.
func (c *Client) receiveRecords(limit int) ([]firetail.Record, bool) {
	records := []firetail.Record{}
	for {
		select {
		case record, open := <-c.recordsChannel:
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
