package domain

import (
	"fmt"
	"hash/fnv"
)

func hashFileName(id int, s string) string {
	h := fnv.New64a()
	_, err := h.Write([]byte(fmt.Sprintf("%d:%s", id, s)))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum64())
}
