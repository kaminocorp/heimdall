package logs

import "context"

// Syslog implements log ingestion via syslog protocol.
type Syslog struct {
	// TODO: add listener config
}

func NewSyslog() *Syslog {
	return &Syslog{}
}

func (s *Syslog) Connect(ctx context.Context) error {
	return nil
}

func (s *Syslog) Health(ctx context.Context) error {
	return nil
}

func (s *Syslog) Stream(ctx context.Context, out chan<- []byte) error {
	// TODO: listen for syslog messages
	return nil
}

func (s *Syslog) Close() error {
	return nil
}
