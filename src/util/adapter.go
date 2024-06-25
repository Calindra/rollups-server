package util

import "fmt"

func RemoveSelector(payload string) string {
	return fmt.Sprintf("0x%s", payload[10:])
}
