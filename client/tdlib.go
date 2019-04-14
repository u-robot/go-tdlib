package client

/*
#cgo windows CFLAGS: -Ic:/tdlib/v1.3.0/Debug/include
#cgo windows LDFLAGS: -Lc:/tdlib/v1.3.0/Debug/bin -ltdjson
#include <stdlib.h>
#include <td/telegram/td_json_client.h>
#include <td/telegram/td_log.h>
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

// TDClient wraps TDLib client.
type TDClient struct {
	locker     *sync.Mutex
	jsonClient unsafe.Pointer
}

// NewTDClient creates new TDLib client.
func NewTDClient() *TDClient {
	return &TDClient{
		locker:     &sync.Mutex{},
		jsonClient: C.td_json_client_create(),
	}
}

// Send sends request to the TDLib client. May be called from any thread.
func (c *TDClient) Send(request Request) {
	data, _ := json.Marshal(request)

	query := C.CString(string(data))
	defer C.free(unsafe.Pointer(query))

	C.td_json_client_send(c.jsonClient, query)
}

// Receive receives incoming updates and request responses from the TDLib client. May be called
// from any thread, but shouldn't be called simultaneously from two different threads.
// Returned pointer will be deallocated by TDLib during next call to td_json_client_receive
// or td_json_client_execute in the same thread, so it can't be used after that.
func (c *TDClient) Receive(timeout time.Duration) (*Response, error) {
	// To workaround multithreading problems mutex is locked and is not unlocked at the end of this function
	// It is responsibility of the function caller to provide result of handling response to Response.Handled
	// channel when the response will not be used anymore and can be released
	c.Lock("Receive")

	// Wait and receive next event from TDLib client
	result := C.td_json_client_receive(c.jsonClient, C.double(float64(timeout)/float64(time.Second)))
	if result == nil {
		c.Unlock("Receive failed")
		return nil, errors.New("update receiving timeout")
	}

	// Raw JSON data containing in a response
	data := []byte(C.GoString(result))

	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		c.Unlock("Receive unable to unmarshal response")
		return nil, err
	}

	response.Data = data
	response.Handled = make(chan error)

	// Block function finishing until response is handled
	defer func(response *Response) {
		err := <-response.Handled
		if err != nil {
			log.Printf("Error occuried during handling update received from TDLib: %s\n", err)
		}
		c.Unlock("Receive")
	}(&response)

	return &response, nil
}

// Execute synchronously executes TDLib request. May be called from any thread.
// Only a few requests can be executed synchronously.
// Returned pointer will be deallocated by TDLib during next call to td_json_client_receive or
// td_json_client_execute in the same thread, so it can't be used after that.
func (c *TDClient) Execute(request Request) (*Response, error) {
	// To workaround multithreading problems mutex is locked and is not unlocked at the end of this function
	// It is responsibility of the function caller to provide result of handling response to Response.Handled
	// channel when the response will not be used anymore and can be released
	c.Lock("Execute")

	data, _ := json.Marshal(request)

	query := C.CString(string(data))
	defer C.free(unsafe.Pointer(query))

	result := C.td_json_client_execute(c.jsonClient, query)
	if result == nil {
		c.Unlock("Execute failed")
		return nil, errors.New("request can't be parsed")
	}

	data = []byte(C.GoString(result))

	var response Response

	err := json.Unmarshal(data, &response)
	if err != nil {
		c.Unlock("Execute unable to unmarshal response")
		return nil, err
	}

	response.Data = data
	response.Handled = make(chan error)

	// Block function finishing until response is handled
	defer func(response *Response) {
		err := <-response.Handled
		if err != nil {
			log.Printf("Error occuried during handling update received from TDLib: %s\n", err)
		}
		c.Unlock("Execute")
	}(&response)

	return &response, nil
}

// Destroy destroys the TDLib client instance. After this is called the client instance shouldn't be used anymore.
func (c *TDClient) Destroy() {
	C.td_json_client_destroy(c.jsonClient)
}

// Lock locks internal mutex.
func (c *TDClient) Lock(str string) {
	log.Printf("Try lock %p: %x on %s!\n", c.locker, c.locker, str)
	c.locker.Lock()
	log.Printf("Lock %p: %x on %s!\n", c.locker, c.locker, str)
	// debug.PrintStack()
}

// Unlock unlocks internal mutex.
func (c *TDClient) Unlock(str string) {
	// log.Printf("Unlock %p: %x on %s!\n", c.locker, c.locker, str)
	// debug.PrintStack()
	c.locker.Unlock()
}

// SetLogFilePath sets the path to the file where the internal TDLib log will be written.
// By default TDLib writes logs to stderr or an OS specific log.
// Use this function to write the log to a file instead.
func SetLogFilePath(filePath string) {
	query := C.CString(filePath)
	defer C.free(unsafe.Pointer(query))

	C.td_set_log_file_path(query)
}

// SetLogMaxFileSize sets maximum size of the file to where the internal TDLib log is written before the file will be auto-rotated.
// Unused if log is not written to a file. Defaults to 10 MB.
func SetLogMaxFileSize(maxFileSize int64) {
	C.td_set_log_max_file_size(C.longlong(maxFileSize))
}

// SetLogVerbosityLevel sets the verbosity level of the internal logging of TDLib.
// By default the TDLib uses a log verbosity level of 5
func SetLogVerbosityLevel(newVerbosityLevel int) {
	C.td_set_log_verbosity_level(C.int(newVerbosityLevel))
}

type meta struct {
	Type  string `json:"@type"`
	Extra string `json:"@extra"`
}

// Request is a common structure which is base of all requests to TDLib client.
type Request struct {
	meta
	Data map[string]interface{}
}

// MarshalJSON returns Request object as the JSON encoding of Request.
func (request Request) MarshalJSON() ([]byte, error) {
	request.Data["@type"] = request.Type
	request.Data["@extra"] = request.Extra

	return json.Marshal(request.Data)
}

// Response is a common structure which is base of all successful responses from TDLib client.
type Response struct {
	meta
	Data    json.RawMessage
	Handled chan error
}

// NotifyHandled is a function which must be called when response is handled and it will not be used anymore.
func (r *Response) NotifyHandled(err error) {
	if r.Handled != nil {
		r.Handled <- err
	}
}

// ResponseError is a common structure which is base of all failed responses from TDLib client.
type ResponseError struct {
	Err *Error
}

// Error returns string describing reason of TDLib client fail.
func (responseError ResponseError) Error() string {
	return fmt.Sprintf("%d %s", responseError.Err.Code, responseError.Err.Message)
}

func buildResponseError(data json.RawMessage) error {
	respErr, err := UnmarshalError(data)
	if err != nil {
		return err
	}

	return ResponseError{
		Err: respErr,
	}
}

// Int64JSON alias for int64, in order to deal with JSON big number problem.
type Int64JSON int64

// MarshalJSON returns Int64JSON object as the JSON encoding of Int64JSON.
func (v *Int64JSON) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(*v), 10)), nil
}

// UnmarshalJSON sets Int64JSON object to a copy of JSON encoding of Int64JSON.
func (v *Int64JSON) UnmarshalJSON(data []byte) error {
	jsonBigInt, err := strconv.ParseInt(string(data[1:len(data)-1]), 10, 64)
	if err != nil {
		return err
	}

	*v = Int64JSON(jsonBigInt)

	return nil
}

// Type is base type interface.
type Type interface {
	GetType() string
	GetClass() string
}
