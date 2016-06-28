package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type JsonDecodeError struct {
	err  error
	keys []string
}

func (j *JsonDecodeError) Error() string {
	keys := strings.Join(j.keys, ".")
	return fmt.Sprintf("Decode %s failed:%s", keys, j.err.Error())
}

func JsonError(err error, keys ...string) error {
	if err == nil {
		return nil
	}
	return &JsonDecodeError{
		err:  err,
		keys: keys,
	}
}

func JSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)
	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}
