package sqldbs

type Notification struct {
	PID     uint32
	Channel string
	Payload string
}
