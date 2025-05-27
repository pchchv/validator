package validator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/sha3"
)

var (
	oneofValsCache       = map[string][]string{}
	oneofValsCacheRWLock = sync.RWMutex{}
	restrictedTags       = map[string]struct{}{
		diveTag:           {},
		keysTag:           {},
		endKeysTag:        {},
		structOnlyTag:     {},
		omitzero:          {},
		omitempty:         {},
		omitnil:           {},
		skipValidationTag: {},
		utf8HexComma:      {},
		utf8Pipe:          {},
		noStructLevelTag:  {},
		requiredTag:       {},
		isdefault:         {},
	}
	// bakedInAliases is a default mapping of a single validation tag that
	// defines a common or complex set of validation(s) to simplify adding validation to structs
	bakedInAliases = map[string]string{
		"iscolor":         "hexcolor|rgb|rgba|hsl|hsla",
		"country_code":    "iso3166_1_alpha2|iso3166_1_alpha3|iso3166_1_alpha_numeric",
		"eu_country_code": "iso3166_1_alpha2_eu|iso3166_1_alpha3_eu|iso3166_1_alpha_numeric_eu",
	}
)

// Func accepts a FieldLevel interface for all validation needs.
// Return value should be true when validation succeeds.
type Func func(fl FieldLevel) bool

// FuncCtx accepts a context.Context and FieldLevel interface for all validation needs.
// The return value should be true when validation succeeds.
type FuncCtx func(ctx context.Context, fl FieldLevel) bool

// wrapFunc wraps normal Func makes it compatible with FuncCtx.
func wrapFunc(fn Func) FuncCtx {
	if fn == nil {
		return nil
	}

	return func(ctx context.Context, fl FieldLevel) bool {
		return fn(fl)
	}
}

// hasMultiByteCharacter is the validation function for validating if the
// field's value has a multi byte character.
func hasMultiByteCharacter(fl FieldLevel) bool {
	field := fl.Field()
	if field.Len() == 0 {
		return true
	}

	return multibyteRegex().MatchString(field.String())
}

func parseOneOfParam(s string) []string {
	oneofValsCacheRWLock.RLock()
	vals, ok := oneofValsCache[s]
	oneofValsCacheRWLock.RUnlock()
	if !ok {
		oneofValsCacheRWLock.Lock()
		vals = splitParamsRegex().FindAllString(s, -1)
		for i := 0; i < len(vals); i++ {
			vals[i] = strings.ReplaceAll(vals[i], "'", "")
		}

		oneofValsCache[s] = vals
		oneofValsCacheRWLock.Unlock()
	}

	return vals
}

// requireCheckFieldValue is a func for check field value
func requireCheckFieldValue(fl FieldLevel, param, value string, defaultNotFoundValue bool) bool {
	field, kind, _, found := fl.GetStructFieldOKAdvanced(fl.Parent(), param)
	if !found {
		return defaultNotFoundValue
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == asInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint() == asUint(value)
	case reflect.Float32:
		return field.Float() == asFloat32(value)
	case reflect.Float64:
		return field.Float() == asFloat64(value)
	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(field.Len()) == asInt(value)
	case reflect.Bool:
		return field.Bool() == (value == "true")
	case reflect.Ptr:
		if field.IsNil() {
			return value == "nil"
		}

		// handle non-nil pointers
		return requireCheckFieldValue(fl, param, value, defaultNotFoundValue)
	}

	// default reflect.String:
	return field.String() == value
}

// requiredIf is the validation function.
// The field under validation must be present and not empty only if all the
// other specified fields are equal to the value following with the specified field.
func requiredIf(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for required_if %s", fl.FieldName()))
	}

	for i := 0; i < len(params); i += 2 {
		if !requireCheckFieldValue(fl, params[i], params[i+1], false) {
			return true
		}
	}

	return hasValue(fl)
}

// requireCheckFieldKind is a func for check field kind.
func requireCheckFieldKind(fl FieldLevel, param string, defaultNotFoundValue bool) bool {
	var nullable, found bool
	field := fl.Field()
	kind := field.Kind()
	if len(param) > 0 {
		field, kind, nullable, found = fl.GetStructFieldOKAdvanced(fl.Parent(), param)
		if !found {
			return defaultNotFoundValue
		}
	}

	switch kind {
	case reflect.Invalid:
		return defaultNotFoundValue
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return field.IsNil()
	default:
		if nullable && field.Interface() != nil {
			return false
		}

		return field.IsValid() && field.IsZero()
	}
}

// requiredUnless is the validation function.
// The field under validation must be present and not empty only unless all the
// other specified fields are equal to the value following with the specified field.
func requiredUnless(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for required_unless %s", fl.FieldName()))
	}

	for i := 0; i < len(params); i += 2 {
		if requireCheckFieldValue(fl, params[i], params[i+1], false) {
			return true
		}
	}

	return hasValue(fl)
}

// requiredWith is the validation function.
// The field under validation must be present and not empty only if any of the
// other specified fields are present.
func requiredWith(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	for _, param := range params {
		if !requireCheckFieldKind(fl, param, true) {
			return hasValue(fl)
		}
	}
	return true
}

// requiredWithAll is the validation function.
// The field under validation must be present and not empty only if all of the
// other specified fields are present.
func requiredWithAll(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	for _, param := range params {
		if requireCheckFieldKind(fl, param, true) {
			return true
		}
	}
	return hasValue(fl)
}

// requiredWithout is the validation function.
// The field under validation must be present and not empty only when any of the
// other specified fields are not present.
func requiredWithout(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	for _, param := range params {
		if requireCheckFieldKind(fl, param, true) {
			return hasValue(fl)
		}
	}
	return true
}

// requiredWithoutAll is the validation function.
// The field under validation must be present and not empty only when all of the
// other specified fields are not present.
func requiredWithoutAll(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	for _, param := range params {
		if !requireCheckFieldKind(fl, param, true) {
			return true
		}
	}
	return hasValue(fl)
}

// digitsHaveLuhnChecksum returns true if and only if the last element of the
// given digits slice is the Luhn checksum of the previous elements.
func digitsHaveLuhnChecksum(digits []string) bool {
	var sum int
	size := len(digits)
	for i, digit := range digits {
		value, err := strconv.Atoi(digit)
		if err != nil {
			return false
		}

		if size%2 == 0 && i%2 == 0 || size%2 == 1 && i%2 == 1 {
			v := value * 2
			if v >= 10 {
				sum += 1 + (v % 10)
			} else {
				sum += v
			}
		} else {
			sum += value
		}
	}
	return (sum % 10) == 0
}

// skipUnless is the validation function.
// The field under validation must be present and not empty only unless all the
// other specified fields are equal to the value following with the specified field.
func skipUnless(fl FieldLevel) bool {
	params := parseOneOfParam(fl.Param())
	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for skip_unless %s", fl.FieldName()))
	}

	for i := 0; i < len(params); i += 2 {
		if !requireCheckFieldValue(fl, params[i], params[i+1], false) {
			return true
		}
	}

	return hasValue(fl)
}

// hasLuhnChecksum is the validation for validating if the current field's value has a valid Luhn checksum.
func hasLuhnChecksum(fl FieldLevel) bool {
	field := fl.Field()
	var str string // convert to a string which will then be split into single digits; easier and more readable than shifting/extracting single digits from a number
	switch field.Kind() {
	case reflect.String:
		str = field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(field.Uint(), 10)
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}

	size := len(str)
	if size < 2 { // there has to be at least one digit that carries a meaning + the checksum
		return false
	}

	digits := strings.Split(str, "")
	return digitsHaveLuhnChecksum(digits)
}

func isOneOf(fl FieldLevel) bool {
	var v string
	vals := parseOneOfParam(fl.Param())
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		v = field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = strconv.FormatUint(field.Uint(), 10)
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}

	for i := 0; i < len(vals); i++ {
		if vals[i] == v {
			return true
		}
	}

	return false
}

// isOneOfCI is the validation function for validating if the
// current field's value is one of the provided string values
// (case insensitive).
func isOneOfCI(fl FieldLevel) bool {
	vals := parseOneOfParam(fl.Param())
	field := fl.Field()
	if field.Kind() != reflect.String {
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}

	v := field.String()
	for _, val := range vals {
		if strings.EqualFold(val, v) {
			return true
		}
	}

	return false
}

func isHTML(fl FieldLevel) bool {
	return hTMLRegex().MatchString(fl.Field().String())
}

func isURLEncoded(fl FieldLevel) bool {
	return uRLEncodedRegex().MatchString(fl.Field().String())
}

func isHTMLEncoded(fl FieldLevel) bool {
	return hTMLEncodedRegex().MatchString(fl.Field().String())
}

// isIP is the validation function for validating if the
// field's value is a valid v4 or v6 IP address.
func isIP(fl FieldLevel) bool {
	ip := net.ParseIP(fl.Field().String())
	return ip != nil
}

// isIPv4 is the validation function for validating if a value is a valid v4 IP address.
func isIPv4(fl FieldLevel) bool {
	ip := net.ParseIP(fl.Field().String())
	return ip != nil && ip.To4() != nil
}

// isIPv6 is the validation function for validating if the
// field's value is a valid v6 IP address.
func isIPv6(fl FieldLevel) bool {
	ip := net.ParseIP(fl.Field().String())
	return ip != nil && ip.To4() == nil
}

// isCIDR is the validation function for validating if the
// field's value is a valid v4 or v6 CIDR address.
func isCIDR(fl FieldLevel) bool {
	_, _, err := net.ParseCIDR(fl.Field().String())
	return err == nil
}

// isCIDRv4 is the validation function for validating if the
// field's value is a valid v4 CIDR address.
func isCIDRv4(fl FieldLevel) bool {
	ip, net, err := net.ParseCIDR(fl.Field().String())
	return err == nil && ip.To4() != nil && net.IP.Equal(ip)
}

// isCIDRv6 is the validation function for validating if the
// field's value is a valid v6 CIDR address.
func isCIDRv6(fl FieldLevel) bool {
	ip, _, err := net.ParseCIDR(fl.Field().String())
	return err == nil && ip.To4() == nil
}

// isMAC is the validation function for validating if the
// field's value is a valid MAC address.
func isMAC(fl FieldLevel) bool {
	_, err := net.ParseMAC(fl.Field().String())
	return err == nil
}

// isSSN is the validation function for validating if the
// field's value is a valid SSN.
func isSSN(fl FieldLevel) bool {
	field := fl.Field()
	if field.Len() != 11 {
		return false
	}

	return sSNRegex().MatchString(field.String())
}

// isUnique is the validation function for validating if each array|slice|map value is unique
func isUnique(fl FieldLevel) bool {
	field := fl.Field()
	param := fl.Param()
	v := reflect.ValueOf(struct{}{})
	switch field.Kind() {
	case reflect.Slice, reflect.Array:
		elem := field.Type().Elem()
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		if param == "" {
			m := reflect.MakeMap(reflect.MapOf(elem, v.Type()))
			for i := 0; i < field.Len(); i++ {
				m.SetMapIndex(reflect.Indirect(field.Index(i)), v)
			}

			return field.Len() == m.Len()
		}

		sf, ok := elem.FieldByName(param)
		if !ok {
			panic(fmt.Sprintf("Bad field name %s", param))
		}

		sfTyp := sf.Type
		if sfTyp.Kind() == reflect.Ptr {
			sfTyp = sfTyp.Elem()
		}

		var fieldlen int
		m := reflect.MakeMap(reflect.MapOf(sfTyp, v.Type()))
		for i := 0; i < field.Len(); i++ {
			key := reflect.Indirect(reflect.Indirect(field.Index(i)).FieldByName(param))
			if key.IsValid() {
				fieldlen++
				m.SetMapIndex(key, v)
			}
		}

		return fieldlen == m.Len()
	case reflect.Map:
		var m reflect.Value
		if field.Type().Elem().Kind() == reflect.Ptr {
			m = reflect.MakeMap(reflect.MapOf(field.Type().Elem().Elem(), v.Type()))
		} else {
			m = reflect.MakeMap(reflect.MapOf(field.Type().Elem(), v.Type()))
		}

		for _, k := range field.MapKeys() {
			m.SetMapIndex(reflect.Indirect(field.MapIndex(k)), v)
		}

		return field.Len() == m.Len()
	default:
		if parent := fl.Parent(); parent.Kind() == reflect.Struct {
			uniqueField := parent.FieldByName(param)
			if uniqueField == reflect.ValueOf(nil) {
				panic(fmt.Sprintf("Bad field name provided %s", param))
			}

			if uniqueField.Kind() != field.Kind() {
				panic(fmt.Sprintf("Bad field type %T:%T", field.Interface(), uniqueField.Interface()))
			}

			return field.Interface() != uniqueField.Interface()
		}

		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}
}

// isLongitude is the validation function for validating if the field's value is a valid longitude coordinate.
func isLongitude(fl FieldLevel) bool {
	var v string
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		v = field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32:
		v = strconv.FormatFloat(field.Float(), 'f', -1, 32)
	case reflect.Float64:
		v = strconv.FormatFloat(field.Float(), 'f', -1, 64)
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}

	return longitudeRegex().MatchString(v)
}

// isLatitude is the validation function for validating if the field's value is a valid latitude coordinate.
func isLatitude(fl FieldLevel) bool {
	var v string
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		v = field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32:
		v = strconv.FormatFloat(field.Float(), 'f', -1, 32)
	case reflect.Float64:
		v = strconv.FormatFloat(field.Float(), 'f', -1, 64)
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}

	return latitudeRegex().MatchString(v)
}

// isDataURI is the validation function for validating if the
// field's value is a valid data URI.
func isDataURI(fl FieldLevel) bool {
	uri := strings.SplitN(fl.Field().String(), ",", 2)
	if len(uri) != 2 || !dataURIRegex().MatchString(uri[0]) {
		return false
	}

	return base64Regex().MatchString(uri[1])
}

// isASCII is the validation function for validating if the
// field's value is a valid ASCII character.
func isASCII(fl FieldLevel) bool {
	return aSCIIRegex().MatchString(fl.Field().String())
}

// isPrintableASCII is the validation function for validating if the
// field's value is a valid printable ASCII character.
func isPrintableASCII(fl FieldLevel) bool {
	return printableASCIIRegex().MatchString(fl.Field().String())
}

// isUUID is the validation function for validating if the
// field's value is a valid UUID of any version.
func isUUID(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uUIDRegex, fl)
}

// isUUID3 is the validation function for validating if the
// field's value is a valid v3 UUID.
func isUUID3(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uUID3Regex, fl)
}

// isUUID4 is the validation function for validating if the
// field's value is a valid v4 UUID.
func isUUID4(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uUID4Regex, fl)
}

// isUUID5 is the validation function for validating if the
// field's value is a valid v5 UUID.
func isUUID5(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uUID5Regex, fl)
}

// isUUIDRFC4122 is the validation function for validating if the
// field's value is a valid RFC4122 UUID of any version.
func isUUIDRFC4122(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uUIDRFC4122Regex, fl)
}

// isUUID3RFC4122 is the validation function for validating if the
// field's value is a valid RFC4122 v3 UUID.
func isUUID3RFC4122(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uUID3RFC4122Regex, fl)
}

// isUUID4RFC4122 is the validation function for validating if the
// field's value is a valid RFC4122 v4 UUID.
func isUUID4RFC4122(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uUID4RFC4122Regex, fl)
}

// isUUID5RFC4122 is the validation function for validating if the
// field's value is a valid RFC4122 v5 UUID.
func isUUID5RFC4122(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uUID5RFC4122Regex, fl)
}

// isULID is the validation function for validating if the
// field's value is a valid ULID.
func isULID(fl FieldLevel) bool {
	return fieldMatchesRegexByStringerValOrString(uLIDRegex, fl)
}

// isSHA256 is the validation function for validating if the field's value is a valid SHA256.
func isSHA256(fl FieldLevel) bool {
	return sha256Regex().MatchString(fl.Field().String())
}

// isSHA384 is the validation function for validating if the field's value is a valid SHA384.
func isSHA384(fl FieldLevel) bool {
	return sha384Regex().MatchString(fl.Field().String())
}

// isSHA512 is the validation function for validating if the field's value is a valid SHA512.
func isSHA512(fl FieldLevel) bool {
	return sha512Regex().MatchString(fl.Field().String())
}

// isMD4 is the validation function for validating if the field's value is a valid MD4.
func isMD4(fl FieldLevel) bool {
	return md4Regex().MatchString(fl.Field().String())
}

// isMD5 is the validation function for validating if the field's value is a valid MD5.
func isMD5(fl FieldLevel) bool {
	return md5Regex().MatchString(fl.Field().String())
}

// isRIPEMD128 is the validation function for validating if the
// field's value is a valid PIPEMD128.
func isRIPEMD128(fl FieldLevel) bool {
	return ripemd128Regex().MatchString(fl.Field().String())
}

// isRIPEMD160 is the validation function for validating if the
// field's value is a valid PIPEMD160.
func isRIPEMD160(fl FieldLevel) bool {
	return ripemd160Regex().MatchString(fl.Field().String())
}

// isTIGER128 is the validation function for validating if the
// field's value is a valid TIGER128.
func isTIGER128(fl FieldLevel) bool {
	return tiger128Regex().MatchString(fl.Field().String())
}

// isTIGER160 is the validation function for validating if the
// field's value is a valid TIGER160.
func isTIGER160(fl FieldLevel) bool {
	return tiger160Regex().MatchString(fl.Field().String())
}

// isTIGER192 is the validation function for validating if the
// field's value is a valid isTIGER192.
func isTIGER192(fl FieldLevel) bool {
	return tiger192Regex().MatchString(fl.Field().String())
}

// isISBN is the validation function for validating if the
// field's value is a valid v10 or v13 ISBN.
func isISBN(fl FieldLevel) bool {
	return isISBN10(fl) || isISBN13(fl)
}

// isISBN10 is the validation function for validating if the
// field's value is a valid v10 ISBN.
func isISBN10(fl FieldLevel) bool {
	s := strings.Replace(strings.Replace(fl.Field().String(), "-", "", 3), " ", "", 3)
	if !iSBN10Regex().MatchString(s) {
		return false
	}

	var i, checksum int32
	for i = 0; i < 9; i++ {
		checksum += (i + 1) * int32(s[i]-'0')
	}

	if s[9] == 'X' {
		checksum += 10 * 10
	} else {
		checksum += 10 * int32(s[9]-'0')
	}

	return checksum%11 == 0
}

// isISBN13 is the validation function for validating if the field's value is a valid v13 ISBN.
func isISBN13(fl FieldLevel) bool {
	s := strings.Replace(strings.Replace(fl.Field().String(), "-", "", 4), " ", "", 4)
	if !iSBN13Regex().MatchString(s) {
		return false
	}

	var i, checksum int32
	factor := []int32{1, 3}
	for i = 0; i < 12; i++ {
		checksum += factor[i%2] * int32(s[i]-'0')
	}

	return (int32(s[12]-'0'))-((10-(checksum%10))%10) == 0
}

// isISSN is the validation function for validating if the
// field's value is a valid ISSN.
func isISSN(fl FieldLevel) bool {
	s := fl.Field().String()
	if !iSSNRegex().MatchString(s) {
		return false
	}

	var pos, checksum int
	s = strings.ReplaceAll(s, "-", "")
	for i := 0; i < 7; i++ {
		checksum += pos * int(s[i]-'0')
		pos--
	}

	if s[7] == 'X' {
		checksum += 10
	} else {
		checksum += int(s[7] - '0')
	}

	return checksum%11 == 0
}

// isBitcoinAddress is the validation function for validating if the
// field's value is a valid btc address.
func isBitcoinAddress(fl FieldLevel) bool {
	address := fl.Field().String()
	if !btcAddressRegex().MatchString(address) {
		return false
	}

	decode := [25]byte{}
	alphabet := []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
	for _, n := range []byte(address) {
		d := bytes.IndexByte(alphabet, n)
		for i := 24; i >= 0; i-- {
			d += 58 * int(decode[i])
			decode[i] = byte(d % 256)
			d /= 256
		}
	}

	h := sha256.New()
	_, _ = h.Write(decode[:21])
	d := h.Sum([]byte{})
	h = sha256.New()
	_, _ = h.Write(d)
	validchecksum := [4]byte{}
	computedchecksum := [4]byte{}
	copy(computedchecksum[:], h.Sum(d[:0]))
	copy(validchecksum[:], decode[21:])
	return validchecksum == computedchecksum
}

// isBitcoinBech32Address is the validation function for validating if the
// field's value is a valid bech32 btc address.
func isBitcoinBech32Address(fl FieldLevel) bool {
	address := fl.Field().String()
	if !btcLowerAddressRegexBech32().MatchString(address) && !btcUpperAddressRegexBech32().MatchString(address) {
		return false
	}

	am := len(address) % 8
	if am == 0 || am == 3 || am == 5 {
		return false
	}

	address = strings.ToLower(address)
	alphabet := "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	hr := []int{3, 3, 0, 2, 3} // the human readable part will always be bc
	addr := address[3:]
	dp := make([]int, 0, len(addr))
	for _, c := range addr {
		dp = append(dp, strings.IndexRune(alphabet, c))
	}

	ver := dp[0]
	if ver < 0 || ver > 16 || (ver == 0 && len(address) != 42 && len(address) != 62) {
		return false
	}

	p := 1
	values := append(hr, dp...)
	GEN := []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	for _, v := range values {
		b := p >> 25
		p = (p&0x1ffffff)<<5 ^ v

		for i := 0; i < 5; i++ {
			if (b>>uint(i))&1 == 1 {
				p ^= GEN[i]
			}
		}
	}

	if p != 1 {
		return false
	}

	var acc int
	var sw []int
	b := uint(0)
	mv := (1 << 5) - 1
	for _, v := range dp[1 : len(dp)-6] {
		acc = (acc << 5) | v
		b += 5
		for b >= 8 {
			b -= 8
			sw = append(sw, (acc>>b)&mv)
		}
	}

	if len(sw) < 2 || len(sw) > 40 {
		return false
	}

	return true
}

// isEthereumAddress is the validation function for validating if the
// field's value is a valid Ethereum address.
func isEthereumAddress(fl FieldLevel) bool {
	address := fl.Field().String()
	return ethAddressRegex().MatchString(address)
}

// isEthereumAddressChecksum is the validation function for validating if the
// field's value is a valid checksummed Ethereum address.
func isEthereumAddressChecksum(fl FieldLevel) bool {
	address := fl.Field().String()
	if !ethAddressRegex().MatchString(address) {
		return false
	}

	// checksum validation
	address = address[2:] // skip "0x" prefix
	h := sha3.NewLegacyKeccak256()
	// hash.Hash's io.Writer implementation says it never returns an error
	_, _ = h.Write([]byte(strings.ToLower(address)))
	hash := hex.EncodeToString(h.Sum(nil))
	for i := 0; i < len(address); i++ {
		if address[i] <= '9' { // skip 0-9 digits: they don't have upper/lower-case
			continue
		}

		if hash[i] > '7' && address[i] >= 'a' || hash[i] <= '7' && address[i] <= 'F' {
			return false
		}
	}

	return true
}

// isEq is the validation function for validating if the
// current field's value is equal to the param's value.
func isEq(fl FieldLevel) bool {
	field := fl.Field()
	param := fl.Param()
	switch field.Kind() {
	case reflect.String:
		return field.String() == param
	case reflect.Slice, reflect.Map, reflect.Array:
		p := asInt(param)
		return int64(field.Len()) == p
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p := asIntFromType(field.Type(), param)
		return field.Int() == p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		p := asUint(param)
		return field.Uint() == p
	case reflect.Float32:
		p := asFloat32(param)
		return field.Float() == p
	case reflect.Float64:
		p := asFloat64(param)
		return field.Float() == p
	case reflect.Bool:
		p := asBool(param)
		return field.Bool() == p
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}
}

// isEqIgnoreCase is the validation function for validating if the
// current field's string value is equal to the param's value.
// The comparison is case-insensitive.
func isEqIgnoreCase(fl FieldLevel) bool {
	field := fl.Field()
	param := fl.Param()
	switch field.Kind() {
	case reflect.String:
		return strings.EqualFold(field.String(), param)
	default:
		panic(fmt.Sprintf("Bad field type %T", field.Interface()))
	}
}

// isEqCrossStructField is the validation function for validating that the
// current field's value is equal to the field, within a separate struct,
// specified by the param's value.
func isEqCrossStructField(fl FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()
	topField, topKind, _, ok := fl.GetStructFieldOK()
	if !ok || topKind != kind {
		return false
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return topField.Int() == field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return topField.Uint() == field.Uint()
	case reflect.Float32, reflect.Float64:
		return topField.Float() == field.Float()
	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(topField.Len()) == int64(field.Len())
	case reflect.Bool:
		return topField.Bool() == field.Bool()
	case reflect.Struct:
		fieldType := field.Type()
		if fieldType.ConvertibleTo(timeType) && topField.Type().ConvertibleTo(timeType) {
			t := field.Convert(timeType).Interface().(time.Time)
			fieldTime := topField.Convert(timeType).Interface().(time.Time)
			return fieldTime.Equal(t)
		}

		// not Same underlying type i. e. struct and time
		if fieldType != topField.Type() {
			return false
		}
	}

	// default reflect.String:
	return topField.String() == field.String()
}

// isEqField is the validation function for validating if the
// current field's value is equal to the
// field specified by the param's value.
func isEqField(fl FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()
	currentField, currentKind, _, ok := fl.GetStructFieldOK()
	if !ok || currentKind != kind {
		return false
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == currentField.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint() == currentField.Uint()
	case reflect.Float32, reflect.Float64:
		return field.Float() == currentField.Float()
	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(field.Len()) == int64(currentField.Len())
	case reflect.Bool:
		return field.Bool() == currentField.Bool()
	case reflect.Struct:
		fieldType := field.Type()
		if fieldType.ConvertibleTo(timeType) && currentField.Type().ConvertibleTo(timeType) {
			t := currentField.Convert(timeType).Interface().(time.Time)
			fieldTime := field.Convert(timeType).Interface().(time.Time)
			return fieldTime.Equal(t)
		}

		// not Same underlying type i.e. struct and time
		if fieldType != currentField.Type() {
			return false
		}
	}

	// default reflect.String:
	return field.String() == currentField.String()
}

// isNe is the validation function for validating that the
// field's value does not equal the provided param value.
func isNe(fl FieldLevel) bool {
	return !isEq(fl)
}

// isNeField is the validation function for validating if the
// current field's value is not equal to the
// field specified by the param's value.
func isNeField(fl FieldLevel) bool {
	field := fl.Field()
	kind := field.Kind()
	currentField, currentKind, _, ok := fl.GetStructFieldOK()
	if !ok || currentKind != kind {
		return true
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() != currentField.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint() != currentField.Uint()
	case reflect.Float32, reflect.Float64:
		return field.Float() != currentField.Float()
	case reflect.Slice, reflect.Map, reflect.Array:
		return int64(field.Len()) != int64(currentField.Len())
	case reflect.Bool:
		return field.Bool() != currentField.Bool()
	case reflect.Struct:
		fieldType := field.Type()
		if fieldType.ConvertibleTo(timeType) && currentField.Type().ConvertibleTo(timeType) {
			t := currentField.Interface().(time.Time)
			fieldTime := field.Interface().(time.Time)
			return !fieldTime.Equal(t)
		}

		// Not Same underlying type i.e. struct and time
		if fieldType != currentField.Type() {
			return true
		}
	}

	// default reflect.String:
	return field.String() != currentField.String()
}

// hasValue is the validation function for validating if the current field's value is not the default static value.
func hasValue(fl FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		if fl.(*validate).fldIsPointer && field.Interface() != nil {
			return true
		}
		return field.IsValid() && !field.IsZero()
	}
}

// hasNotZeroValue is the validation function for validating if the current field's value is not the zero value for its type.
func hasNotZeroValue(fl FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	default:
		if fl.(*validate).fldIsPointer && field.Interface() != nil {
			return !field.IsZero()
		}
		return field.IsValid() && !field.IsZero()
	}
}
