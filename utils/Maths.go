package utils

var Alphabet = "abcdefghijklmnopqrstuvwxyz"
var Number = "0123456789"

func Product(wordlist, result []string, depth, maxdepth int, keep bool) []string {
	if maxdepth == 1 {
		return wordlist
	}
	if depth == 1 {
		result = wordlist
	}
	if depth < maxdepth {
		var tmp []string
		for _, r := range result {
			for _, word := range wordlist {
				tmp = append(tmp, r+word)
			}
		}
		if keep {
			result = append(result, Product(wordlist, tmp, depth+1, maxdepth, keep)...)
		} else {
			result = Product(wordlist, tmp, depth+1, maxdepth, keep)
		}
	}
	return result
}

func Product2(wordlist, result []string, maxdepth int) []string {
	if maxdepth == 1 {
		return result
	}

	var tmp []string
	for _, r := range result {
		for _, word := range wordlist {
			tmp = append(tmp, r+word)
		}
	}
	return Product2(wordlist, tmp, maxdepth-1)
}
