package proto

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type ErrOffOutOfRange struct {
	Offset uint64
}

func (e ErrOffOutOfRange) GRPCStatus() *status.Status {
	st := status.New(
		404, fmt.Sprintf("offset out of range: %d", e.Offset),
	)
	msg := fmt.Sprintf(
		"The requested offset is outside the log's range: %d",
		e.Offset,
	)

	d := &errdetails.LocalizedMessage{
		Locale:  "en-US",
		Message: msg,
	}
	stwd, err := st.WithDetails(d)
	if err != nil {
		return st
	}
	return stwd
}

func (e ErrOffOutOfRange) Error() string {
	return e.GRPCStatus().Err().Error()
}
