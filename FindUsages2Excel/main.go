package FindUsages2Excel

import (
	"os"
	"bufio"
	"regexp"
	"strings"
	"github.com/tealeg/xlsx"
)

type token struct {
	mapData map[string]*token
	parent *token
}

func main() {
	writeNewFile(scanFile("usagesCategory.txt"), "v2/grouped_cats.xlsx")
	writeNewFile(scanFile("usagesParamsValues.txt"), "v2/grouped_params.xlsx")
	writeNewFile(scanFile("usagesAll.txt"), "v2/grouped_all.xlsx")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func scanFile(filename string) []string {
	file, err:= os.Open(filename)
	check(err)
	defer file.Close()

	data := token{make(map[string]*token), nil}
	spacesRegexp, _ := regexp.Compile("^\\s*")
	scanner := bufio.NewScanner(file)
	caret := &data
	prev := 0
	strPrev := ""

	// WARNING!! Scanner does not deal well with lines longer than 65536 characters
	for scanner.Scan() {
		curr := spacesRegexp.FindStringIndex(scanner.Text())[1] / 4

		if curr - prev > 1 {
			panic("Invalid string tabs")
		}

		str := strings.TrimSpace(scanner.Text())

		if prev > curr {
			// Jump out group
			for i := prev - curr; i > 0; i-- {
				caret = caret.parent
			}
		}

		if curr - prev == 1 {
			// Dive into group
			caret = caret.mapData[strPrev]
		}

		if _, ok := caret.mapData[str]; !ok {
			caret.mapData[str] = &(token{map[string]*token{}, caret})
		}

		prev = curr
		strPrev = str
	}

	check(scanner.Err())

	rows := []string{}
	expandData(data, 0, "", &rows)

	return rows
}

func expandData(data token, level int, prev string, cells *[]string) {
	for k, v := range data.mapData {
		if level < 3 {
			expandData(*v, level + 1, "", cells)
		} else if level == 3 {
			expandData(*v, level + 1, prev + k + "/", cells)
		} else if level == 4 {
			expandData(*v, level + 1, prev + k, cells)
		} else {
			*cells = append(*cells, prev + "::" + k)
		}

	}
}

func writeNewFile(rows []string, filename string) {
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var err error

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("List1")
	check(err)

	for _, totalStr := range rows {
		row := sheet.AddRow()
		for _, cellStr := range processStringData(totalStr) {
			cell := row.AddCell()
			cell.Value = cellStr
		}
	}

	err = file.Save(filename)
	check(err)
}

func processStringData(input string) []string {
	tRegexp, _ := regexp.Compile("^(.+::\\d+)(\\s*)(.*)$")
	tmp := tRegexp.FindStringSubmatch(input)

	usages := findUsages(tmp[3])
	return append([]string{}, tmp[1], tmp[3], usages[0], usages[1])
}

func findUsages(input string) []string {
	ansSlice := []string{}
	ans := []string{""}

	if isUse, _ := regexp.MatchString("(use|class|extends|implements) Obj_\\w+", input); isUse {
		// игнорируем use, class, extends, implements
		input = ""
	} else if isComment, _ := regexp.MatchString("@(var|return|param|see|covers).*Obj_\\w+", input); isComment {
		// игнорируем use, class, extends, implements
		input = ""
	} else if isTypehint, _ := regexp.MatchString("[(\\s,;]?Obj_\\w+\\s?(\\$.+)?$", input); isTypehint {
		// игнорируем Typehint
		input = ""
	} else if isClassGet, _ := regexp.MatchString("Obj_\\w+::class", input); isClassGet {
		// игнорируем Typehint
		input = ""
	} else {
		// Instanceof или new
		iRegexp, _ := regexp.Compile("(instanceof|new) \\\\?Obj_\\w+")
		if iofMatches := iRegexp.FindAllStringSubmatch(input, -1); len(iofMatches) > 0 {
			ansSlice = append(ansSlice, "Конструкция " + iofMatches[0][1])
			input = ""
		}

		// Константы классов
		sRegexp, _ := regexp.Compile("(Obj_\\w+::[A-Z0-9_]+)([\\s+=\\-*/;,:)\\]]+)?")
		if stats := sRegexp.FindAllStringSubmatch(input, -1); len(stats) > 0 {
			for _, usage := range stats {
				input = strings.Replace(input, usage[1], "", -1)
			}

			ansSlice = append(ansSlice, "Константы классов")
		}

		// Статические переменные
		spRegexp, _ := regexp.Compile("Obj_\\w+::\\$\\w+")
		if props := spRegexp.FindAllStringSubmatch(input, -1); len(props) > 0 {
			for _, usage := range props {
				input = strings.Replace(input, usage[0], "", -1)
			}

			ansSlice = append(ansSlice, "Статические переменные")
		}

		// Методы
		mRegexp, _ := regexp.Compile("Obj_\\w+(::|->)\\w+\\(")
		for _, usage := range mRegexp.FindAllStringSubmatch(input, -1) {
			ansSlice = append(ansSlice, usage[0] + ")")
			input = strings.Replace(input, usage[0], "", -1)
		}

		// Array map
		mapRegexp, _ := regexp.Compile("array_map\\(\\['Obj_\\w+'[,$'\\w\\s]+\\]")
		for _, usage := range mapRegexp.FindAllStringSubmatch(input, -1) {
			ansSlice = append(ansSlice, usage[0] + ")")
			input = strings.Replace(input, usage[0], "", -1)
		}
	}

	for i := 0; i < len(ansSlice); i++ {
		ans[0] += ansSlice[i]
		if i < len(ansSlice) - 1 {
			ans[0] += "\n"
		}
	}

	if isEmptyInput, _ := regexp.MatchString("Obj_\\w+", input); isEmptyInput {
		return append(ans, input)
	} else {
		return append(ans, "")
	}
}