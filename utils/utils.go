package utils

func isExistString(strArr []string, str string) bool {
	for _, v := range strArr {
		if v == str {
			return true
		}
	}

	return false
}

func ConvertMapKeysToStrArr(mm map[string]interface{}) []string {
	res := make([]string, 0)

	for k, _ := range mm {
		res = append(res, k)
	}

	return res
}
