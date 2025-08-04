package internal

func ErrorToString(err error) string {
	if nil == err {
		return ""
	}
	return err.Error()
}
