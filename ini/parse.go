package ini

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/zzztttkkk/snow/utils"
)

//noinspection RegExpRedundantEscape
var iniEnvRegexp = regexp.MustCompile(`^\$ENV\{\w+}$`)

//noinspection RegExpRedundantEscape
var iniNameRegexp = regexp.MustCompile(`^\$\{[\w._]+}$`)

func doReplace(config *Config, v []byte, currentSectionName string, currentResult map[string][]byte) []byte {
	if bytes.HasPrefix(v, []byte("file://")) {
		file, err := os.Open(string(v[7:]))
		if err != nil {
			panic(err)
		}
		v, e := ioutil.ReadAll(file)
		if e != nil {
			panic(e)
		}
		return v
	}

	vs := iniEnvRegexp.ReplaceAllFunc(
		v,
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
				_v = config.storage[name]
			}

			if len(_v) < 1 {
				panic(fmt.Errorf("snow.ini: val: `%s` not found", name))
			}
			return _v
		},
	)

	return vs
}

func parseIniFile(config *Config, filename string) {
	var err error

	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	sectionName := ""
	lineNum := 0
	for scanner.Scan() {
		lineNum += 1
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) < 1 || line[0] == '#' || line[0] == ';' {
			continue
		}
		k, sn, v, isValid, isErr := parseIniLine(sectionName, line)
		sectionName = sn

		if !isValid {
			if isErr {
				panic(fmt.Errorf("snow.ini: line: %d", lineNum))
			}
			continue
		}

		_k := k
		if len(sectionName) > 1 {
			_k = sectionName + "." + k
		}

		config.storage[_k] = doReplace(config, v, sectionName, config.storage)
		config.raw[_k] = string(v)
	}

	err = scanner.Err()
	if err != nil {
		panic(err)
	}
}

var iniKeyRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*(\.[a-zA-Z][a-zA-Z0-9_]*)*$`)

func parseIniLine(sectionName string, line []byte) (k, sn string, v []byte, isValid, isErr bool) {
	sn = sectionName

	if line[0] == '[' {
		ind := bytes.IndexByte(line, ']')
		if ind < 0 {
			isErr = true
			return
		}
		if !iniKeyRegexp.Match(line[1:ind]) {
			isErr = true
			return
		}
		sn = string(line[1:ind])
		return
	}

	ind := bytes.IndexByte(line, '=')
	if ind < 1 {
		isErr = true
		return
	}

	k = string(bytes.TrimSpace(line[:ind]))
	v = bytes.TrimSpace(line[ind+1:])
	isValid = true
	return
}
