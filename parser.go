package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"unicode/utf8"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/dlclark/regexp2"
)

func parse(reader io.Reader) string {
	f, err := excelize.OpenReader(reader)

	if err != nil {
		fmt.Println(err)
		return err.Error()
	}

	sheets := f.GetSheetList()

	rows, err := f.GetRows(sheets[0])

	if err != nil {
		fmt.Println("파싱에 실패했습니다")
		return err.Error()
	}

	mealIndexes := [3][2]int{}
	currentStep := 0

	// Find Meal Indexes
	for rowIndex, row := range rows {
		firstChar, _ := utf8.DecodeRuneInString(row[0])
		lastChar, _ := utf8.DecodeLastRuneInString(row[0])

		if lastChar == '식' {
			mealIndexes[currentStep][0] = rowIndex
		} else if firstChar == '영' {
			mealIndexes[currentStep][1] = rowIndex
			currentStep++
		}
	}

	meals := make([][][]string, 7)

	for i := range meals {
		meals[i] = make([][]string, 3)
	}

	reg := regexp2.MustCompile(".*?[가-힣](?=[^가-힣]*?$)", 0)

	// 아침점심저녁
	for timeIndex, rowIndex := range mealIndexes {
		// 각 타임 시작 ~ 끝
		for currentRow := rowIndex[0]; currentRow < rowIndex[1]; currentRow++ {
			for dayIndex, menu := range rows[currentRow] {
				if dayIndex >= 7 || dayIndex == 0 {
					continue
				}

				if len(menu) == 0 {
					continue
				}
				matched, _ := reg.FindStringMatch(menu)

				if matched != nil {
					meals[dayIndex-1][timeIndex] = append(meals[dayIndex-1][timeIndex], matched.String())
				}
			}
		}
	}

	data, _ := json.Marshal(meals)

	return string(data)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(parse(req.Body)))
	})

	fmt.Println("Server is listening on PORT 5000")
	fmt.Println("Request file on root route (/)")

	http.ListenAndServe(":5000", nil)
}
