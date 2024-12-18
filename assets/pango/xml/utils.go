package xml

import "bytes"

func StripEntryElement(data []byte) []byte {
	if !bytes.HasPrefix(data, []byte("<entry")) || !bytes.HasSuffix(data, []byte("</entry>")) {
		return data
	}

	var startIdx, endIdx int
	startIdx = bytes.Index(data, []byte(">"))
	endIdx = len(data) - len("</entry>")

	return data[startIdx+1 : endIdx]
}
