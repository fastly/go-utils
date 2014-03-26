package jlogutil

import (
	"errors"

	"github.com/fastly/jlog-go"
)

// JReader embeds a reader with a subscriber name.
type JReader struct {
	jlog.Reader
	subscriber string
}

// NewReader opens the jlog at the given path, attaches the given subscriber
// for reading and returns the opened jlog.
func NewReader(path, subscriber string) (*JReader, error) {
	log, err := jlog.NewReader(path, nil)
	if err != nil {
		return nil, err
	}
	subs, err := log.ListSubscribers()
	if err != nil {
		return nil, err
	}

	jreader := &JReader{Reader: log}

	for _, sub := range subs {
		if sub == subscriber {
			return nil, errors.New("subscriber already exists")
		}
	}

	err = log.AddSubscriber(subscriber, jlog.END)
	if err != nil {
		return nil, err
	}
	err = log.Open(subscriber)
	if err != nil {
		jreader.Close()
		return nil, err
	}
	jreader.subscriber = subscriber
	return jreader, nil
}

func (j *JReader) Reopen() error {
	j.Reader.Close()
	err := j.Open(j.subscriber)
	if err != nil {
		return err
	}
	return nil
}

// Removes the JReader's specific reader and closes the jlog.
func (j *JReader) Close() {
	j.RemoveSubscriber(j.subscriber)
	j.Reader.Close()
}
