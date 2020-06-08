package ini

import (
	"bufio"
	"fmt"
	"github.com/zzztttkkk/snow/utils"
	"os"
	"regexp"
	"strings"
)

//noinspection RegExpRedundantEscape
var iniEnvRegexp = regexp.MustCompile(`\$ENV\{\w+}`)

//noinspection RegExpRedundantEscape
var iniNameRegexp = regexp.MustCompile(`\$\{[\w._]+}`)

func doReplace(v string, currentSectionName string, currentResult map[string]string) string {
	vs := iniEnvRegexp.ReplaceAllFunc(
		utils.S2b(v),
		func(bytes []byte) []byte {
			name := string(bytes[5 : len(bytes)-1])
			val := os.Getenv(name)
			if len(val) < 1 {
				panic(fmt.Errorf("snow.ini: env: `%s` not found", name))
			}
			return utils.S2b(val)
		},
	)

	vs = iniNameRegexp.ReplaceAllFunc(
		vs,
		func(bytes []byte) []byte {
			name := string(bytes[2 : len(bytes)-1])
			if name[0] == '.' {
				name = currentSectionName + name
			}

			_v, ok := currentResult[name]
			if !ok {
				_v = storage[name]
			}

			if len(_v) < 1 {
				panic(fmt.Errorf("snow.ini: val: `%s` not found", name))
			}
			return utils.S2b(_v)
		},
	)

	return utils.B2s(vs)
}

func parseIniFile(filename string) map[string]string {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		return nil
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	sectionName := ""
	lineNum := 0
	for scanner.Scan() {
		lineNum += 1
		line := strings.TrimSpace(scanner.Text())
		if len(line) < 1 || line[0] == '#' {
			continue
		}
		k, v, sn, isValid, isErr := parseIniLine(sectionName, line)
		sectionName = sn

		if !isValid {
			if isErr {
				err = fmt.Errorf("snow.ini: line: %d", lineNum)
				return nil
			}
			continue
		}

		_k := k
		if len(sectionName) > 1 {
			_k = sectionName + "." + k
		}

		_, exist := result[_k]
		if exist {
			err = fmt.Errorf("snow.ini: duplicate cacheKey `%s`; line: %d", _k, lineNum)
			return nil
		}
		result[_k] = doReplace(v, sectionName, result)
	}

	err = scanner.Err()
	if err != nil {
		return nil
	}
	return result
}

var iniKeyRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z][a-zA-Z0-9_]*)*$`)

func parseIniLine(sectionName, line string) (k, v, sn string, isValid, isErr bool) {
	sn = sectionName

	if line[0] == '[' {
		ind := strings.Index(line, "]")
		if ind < 0 {
			isErr = true
			return
		}
		if !iniKeyRegexp.MatchString(line[1:ind]) {
			isErr = true
			return
		}
		sn = line[1:ind]
		return
	}

	s := strings.Split(line, "=")
	if len(s) < 2 {
		isErr = true
		return
	}

	k = strings.TrimSpace(s[0])
	if !iniKeyRegexp.MatchString(k) {
		isErr = true
		return
	}

	v = strings.TrimSpace(strings.Join(s[1:], "="))

	//noinspection GoPreferNilSlice
	var _v = []rune{}
	var quote rune = 0
	var r rune
	var prev rune = 0
	for _, r = range v {
		// handle quote
		if prev != '\\' {
			if quote != 0 && r == quote {
				_v = append(_v, r)
				break
			}
			if quote == 0 && (r == '"' || r == '\'') {
				quote = r
				_v = append(_v, r)
				prev = r
				continue
			}
			if r == '#' && quote == 0 {
				break
			}
		}

		// handle '\'
		if prev == '\\' {
			_v[len(_v)-1] = r
			prev = r
			if r == '\\' {
				prev = _v[len(_v)-2]
			}
		} else {
			prev = r
			_v = append(_v, r)
		}
	}

	__v := ""
	for _, r := range _v {
		__v += string(r)
	}

	__v = strings.TrimSpace(__v)
	ei := len(__v) - 1

	if __v[0] == '"' {
		if __v[ei] != '"' {
			isErr = true
			return
		}
		v = __v[1:ei]
		isValid = true
		return
	}

	if __v[0] == '\'' {
		if __v[ei] != '\'' {
			isErr = true
			return
		}
		v = __v[1:ei]
		isValid = true
		return
	}

	v = __v
	isValid = true
	return
}
