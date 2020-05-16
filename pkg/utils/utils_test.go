package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestRandString(t *testing.T) {
	var strLen int
	var randStr string
	var exists bool
	rand.Seed(time.Now().UnixNano())
	randStrings := make(map[string]struct{})
	for i := 0; i < 2000; i++ {
		strLen = rand.Intn(20) + 10
		randStr = RandString(strLen)
		assert.Len(t, randStr, strLen)
		_, exists = randStrings[randStr]
		assert.False(t, exists, fmt.Sprintf("not unique value %s on iteration %d", randStr, i))
		if exists {
			break
		}
		randStrings[randStr] = struct{}{}
	}
}

func TestParseInt(t *testing.T) {
	var num int
	var expectedValue int
	var result int
	rand.Seed(time.Now().UnixNano())
	defaultValue, minValue, maxValue := 30, 2, 100
	for i := 0; i < 100; i++ {
		num = rand.Intn(120)
		if num < minValue || num > maxValue {
			expectedValue = defaultValue
		} else {
			expectedValue = num
		}
		result = ParseInt(strconv.Itoa(num), defaultValue, minValue, maxValue)
		assert.Equal(t, expectedValue, result)
	}
}

func TestInArray(t *testing.T) {
	values := []string{"a", "b", "c", "d"}
	for _, v := range values {
		assert.True(t, InArray(values, v))
	}
	for _, iv := range []string{"e", "f", "g", "h"} {
		assert.False(t, InArray(values, iv))
	}
}

func TestIsLengthValid(t *testing.T) {
	var result bool
	result = IsLengthValid("test", 2, 10)
	assert.True(t, result)

	result = IsLengthValid("", 2, 10)
	assert.False(t, result)

	result = IsLengthValid("1234567891011", 2, 10)
	assert.False(t, result)

	result = IsLengthValid("разДваТри!", 2, 10)
	assert.True(t, result)
}

func TestIsEmailValid(t *testing.T) {
	assert.True(t, IsEmailValid("test@mail.com"))
	assert.True(t, IsEmailValid("tes.asdsa.asd-t@mail.com"))
	assert.True(t, IsEmailValid("a@gm.ru"))
	assert.True(t, IsEmailValid("ADSasAS-as._AsdAsl@g.kg"))

	assert.False(t, IsEmailValid("tes t@gmail.com"))
	assert.False(t, IsEmailValid("тест@мейл.рф"))
	assert.False(t, IsEmailValid("test"))
}

func TestIsNameValid(t *testing.T) {
	assert.True(t, IsNameValid("Cheburek"))
	assert.True(t, IsNameValid("Чебурек"))
	assert.True(t, IsNameValid("Чебурек Кек"))
	assert.True(t, IsNameValid("Чебурек: Кек"))
	assert.True(t, IsNameValid("Чебурек_Кек222"))
	assert.True(t, IsNameValid("0900-989"))
	assert.True(t, IsNameValid("Орешек"))
	assert.True(t, IsNameValid("Фундук"))
	assert.True(t, IsNameValid("До ре ми"))

	assert.False(t, IsNameValid("Фундук "))
	assert.False(t, IsNameValid(" Фундук-"))
}

func TestIsUrlValid(t *testing.T) {
	assert.True(t, IsUrlValid("https://stackoverflow.com/questions/3809401/what-is-a-good-regular-expression-to-match-a-url"))
	assert.True(t, IsUrlValid("https://www.youtube.com/watch?v=0QavEsLbjGY"))
	assert.True(t, IsUrlValid("youtube.com/watch?v=0QavEsLbjGY"))

	assert.False(t, IsUrlValid("ftp://test.com"))
}

func TestGetRandomColor(t *testing.T) {
	for i := 0; i < 10; i++ {
		GetRandomColor()
	}
}
