package kai

// pushData is a struct that holds the data to be pushed into the channel
type pushData struct {
	address string
	rectype string
	msg     string
}

// RecorderPush pushes data into the channel
func (b *BTOrdIdx) RecorderPush(address, rectype, msg string) (err error) {
	data := pushData{
		address: address,
		rectype: rectype,
		msg:     msg,
	}
	b.pushCh <- data
	return nil
}

// SelectPush blocks and retrieves data from the channel
func (b *BTOrdIdx) SelectPush() (address, rectype, msg string) {
	data := <-b.pushCh
	return data.address, data.rectype, data.msg
}
