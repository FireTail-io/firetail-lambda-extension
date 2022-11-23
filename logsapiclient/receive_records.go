package logsapiclient

import (
	"firetail-lambda-extension/firetail"
)

// ReceiveRecords returns a slice of firetail Records up to the size of `limit`, and a boolean indicating that the channel
// still has items to be read - it will only be `false` when the channel is closed & empty. It achieves this by continuously
// reading from the log server's recordsChannel until it's empty, or the size limit has been reached.
func (c *Client) ReceiveRecords(limit int) ([]firetail.Record, bool) {
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
