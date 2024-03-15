package translate

// PackageName get package name from Go SDK path.
func PackageName(list []string) string {
	if len(list) == 0 {
		return ""
	}
	return list[len(list)-1]
}
