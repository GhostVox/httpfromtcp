package headers

import (
	"bytes"
	"errors"
	"fmt"
)

type Headers map[string]string

const (
	clrf = "\r\n"
)

func NewHeaders() Headers {
	return make(Headers)
}
func (h Headers) Set(key, value string) {
	if v, ok := h[key]; ok {
		h[key] = v + ", " + value
		return
	}
	h[key] = value
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(clrf))
	// check if the registered nurse is not found
	if idx == -1 {
		return 0, false, nil
	}
	// check if the registered nurse is at the beginning of the data if so we are done
	if idx == 0 {
		return 2, true, nil
	}
	// get the header data
	header := data[:idx]

	// split the header into key and value
	parts := bytes.SplitN(header, []byte(":"), 2)
	// check if the header key is valid
	if bytes.HasSuffix(parts[0], []byte(" ")) {
		return 0, false, errors.New("Invalid spacing")
	}

	stripedKey := bytes.ToLower(bytes.TrimSpace(parts[0]))
	if !checkHeaderKey(stripedKey) {
		return 0, false, errors.New("Invalid header key")
	}
	key := string(stripedKey)
	value := string(bytes.TrimSpace(parts[1]))
	h.Set(key, value)

	return idx + len(clrf), false, nil
}

func checkHeaderKey(key []byte) bool {
	var specialCh = map[byte]bool{
		'!':  true,
		'#':  true,
		'$':  true,
		'%':  true,
		'&':  true,
		'\'': true,
		'*':  true,
		'+':  true,
		'-':  true,
		'.':  true,
		'^':  true,
		'_':  true,
		'`':  true,
		'|':  true,
		'~':  true,
	}
	if len(key) == 0 {
		return false
	}
	for _, b := range key {
		if (b < 'a' || b > 'z') && (b < '0' || b > '9') {

			if _, ok := specialCh[b]; !ok {
				fmt.Printf("b: %c\n", b)
				return false
			}
		}
	}
	return true
}
