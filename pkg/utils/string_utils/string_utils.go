package string_utils

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

//const (
//	layoutDate = "2006-01-02T15:04:05.999"
//)
//
////	Переменная хронящая список параметров для параметризации каждого сервиса
//var searchParams = map[string]list.List{}
//
//	Для генерации используется литерал __rqTm__
//func GetDate() (date string) {
//	date = time.Now().Format(layoutDate)
//	return date
//}
//
//	Для генерации используется литерал __getNewRqUID__
//func GetNewRqUID() (rqUid string) {
//	uuidWithHyphen := uuid.New()
//	rqUid = strings.Replace(uuidWithHyphen.String(), "-", "", -1)
//	zap.S().Debug("NewRqUID:", rqUid)
//	return rqUid
//}
//
//const (
//	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
//	letterIdxBits = 6
//	letterIdxMask = 1<<letterIdxBits - 1
//	letterIdxMax  = 63 / letterIdxBits
//)
//
//var src = rand.NewSource(time.Now().UnixNano())
//
////	Для генерации используется литерал __RandomString(quantity)__
//func RandomString(quantity int) (randomString string) {
//	sb := strings.Builder{}
//	sb.Grow(quantity)
//	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
//	for i, cache, remain := quantity-1, src.Int63(), letterIdxMax; i >= 0; {
//		if remain == 0 {
//			cache, remain = src.Int63(), letterIdxMax
//		}
//		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
//			sb.WriteByte(letterBytes[idx])
//			i--
//		}
//		cache >>= letterIdxBits
//		remain--
//	}
//	randomString = sb.String()
//	zap.S().Debug("RandomString:", randomString)
//	return randomString
//}
//
////	Для генерации используется литерал __RandomInt(quantity)__
////	Возвращается строковое представление числа
//func RandomInt(quantity int) string {
//	atoi, _ := strconv.Atoi(strings.Repeat("9", quantity))
//	intn := rand.Intn(atoi)
//	zap.S().Debug("RandomInt: ", intn)
//	return strconv.Itoa(intn)
//}
//
//func FindStrings(reg *regexp.Regexp, where string) (foundData string) {
//	subMatch := reg.FindStringSubmatch(where)
//	zap.S().Debugf("Submatch: '%v' -> '%v'", reg, subMatch)
//	if len(subMatch) == 2 {
//		foundData = subMatch[1]
//	} else {
//		foundData = ""
//	}
//	zap.S().Infof("FindStrings: '%v'", foundData)
//	return foundData
//}
//

// Маскировка строки для безопастности
func MaskString(src string) string {
	return fmt.Sprint(src[:2], strings.Repeat("*", len(src)-4), src[len(src)-2:])
}

// Вернуть указатель на значение
func ReturnPointer(literal string) (pointer *string) {
	return &literal
}

func StrToInt(i string) (int, error) {
	returnStr, er := strconv.Atoi(i)
	if er != nil {
		zap.S().Warnf("Error converting the string type '%v' in int -> '%v'", i, er)
	}
	return returnStr, er
}
