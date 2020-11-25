package sqlx

type Bytes []byte

var emptyBytes = []byte("\"\"")

func (u Bytes) MarshalJSON() ([]byte, error) {
	if u == nil {
		return emptyBytes, nil
	}

	var ret []byte
	ret = append(ret, '"')

	for _, v := range u {
		switch v {
		case '"':
			ret = append(ret, '\\', v)
		case '\\':
			ret = append(ret, '\\', '\\')
		default:
			ret = append(ret, v)
		}
	}
	ret = append(ret, '"')
	return ret, nil
}
