package utils

import "unicode"

// NaturalLess 实现“大小写不敏感的自然排序”。
//
// 特点：
//   - 数字段按数值比较：file2 < file10（字符串比较会是 file10 < file2）
//   - 字母按大小写折叠比较：Apple 和 apple 视为相同
//   - 支持 Unicode：中文、日文文件名不会乱序
//
// 返回：
//   - -1：a 应排在 b 前面
//   - 0：a 和 b 相等（大小写折叠后）
//   - 1：a 应排在 b 后面
//
// 这是对已失效的 github.com/WithoutPants/sortorder/casefolded 的替代实现。
func NaturalLess(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)

	i, j := 0, 0

	for i < len(ra) && j < len(rb) {
		// 如果两边当前都是数字，进入数字段比较
		if unicode.IsDigit(ra[i]) && unicode.IsDigit(rb[j]) {
			ai, bj := i, j
			for ai < len(ra) && unicode.IsDigit(ra[ai]) {
				ai++
			}
			for bj < len(rb) && unicode.IsDigit(rb[bj]) {
				bj++
			}

			// 跳过前导零，正确处理 "01" 和 "1"
			aNum := trimLeadingZeros(ra[i:ai])
			bNum := trimLeadingZeros(rb[j:bj])

			// 长度不同时，短的小（数值小）
			if len(aNum) != len(bNum) {
				if len(aNum) < len(bNum) {
					return -1
				}
				return 1
			}

			// 长度相同，逐位比较
			for k := 0; k < len(aNum); k++ {
				if aNum[k] != bNum[k] {
					if aNum[k] < bNum[k] {
						return -1
					}
					return 1
				}
			}

			i = ai
			j = bj
			continue
		}

		// 非数字段，按小写 rune 比较
		la := unicode.ToLower(ra[i])
		lb := unicode.ToLower(rb[j])

		if la != lb {
			if la < lb {
				return -1
			}
			return 1
		}

		i++
		j++
	}

	// 谁还有剩余谁就大
	switch {
	case i < len(ra):
		return 1
	case j < len(rb):
		return -1
	default:
		return 0
	}
}

// trimLeadingZeros 去掉数字段的前导零。
func trimLeadingZeros(runes []rune) []rune {
	for len(runes) > 1 && runes[0] == '0' {
		runes = runes[1:]
	}
	return runes
}
