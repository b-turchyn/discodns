package main

import (
	"fmt"

	"github.com/coreos/go-etcd/etcd"
	"github.com/miekg/dns"
)

// NodeConversionError holds the details for when an error occurs converting an etcd node into the expected resource record.
type NodeConversionError struct {
	Message       string
	Node          *etcd.Node
	AttemptedType uint16
}

func (e *NodeConversionError) Error() string {
	return fmt.Sprintf(
		"Unable to convert etc Node into a RR of type %d ('%s'): %s. Node details: %+v",
		e.AttemptedType,
		dns.TypeToString[e.AttemptedType],
		e.Message,
		&e.Node)
}

// RecordValueError holes the error message for when a resource record value was the wrong type.
type RecordValueError struct {
	Message       string
	AttemptedType uint16
}

func (e *RecordValueError) Error() string {
	return fmt.Sprintf(
		"Invalid record value for type %d: %s",
		e.AttemptedType,
		e.Message)
}
