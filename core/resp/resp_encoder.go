package resp

import "fmt"

var (
	RESP_OK = []byte("+OK\r\n")
)

// Encoder encodes responses according to RESP
func Encode(resp interface{}) string {
	switch v := resp.(type) {
	case string:
		return fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
	case error:
		return fmt.Sprintf("-%s\r\n", v.Error())
	case []byte:
		return string(v)
	}

	return ""
}
