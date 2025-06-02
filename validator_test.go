package validator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	. "github.com/pchchv/go-assert"
)

func TestCrossNamespaceFieldValidation(t *testing.T) {
	type SliceStruct struct {
		Name string
	}

	type Inner struct {
		CreatedAt        *time.Time
		Slice            []string
		SliceStructs     []*SliceStruct
		SliceSlice       [][]string
		SliceSliceStruct [][]*SliceStruct
		SliceMap         []map[string]string
		Map              map[string]string
		MapMap           map[string]map[string]string
		MapStructs       map[string]*SliceStruct
		MapMapStruct     map[string]map[string]*SliceStruct
		MapSlice         map[string][]string
		MapInt           map[int]string
		MapInt8          map[int8]string
		MapInt16         map[int16]string
		MapInt32         map[int32]string
		MapInt64         map[int64]string
		MapUint          map[uint]string
		MapUint8         map[uint8]string
		MapUint16        map[uint16]string
		MapUint32        map[uint32]string
		MapUint64        map[uint64]string
		MapFloat32       map[float32]string
		MapFloat64       map[float64]string
		MapBool          map[bool]string
	}

	type Test struct {
		Inner     *Inner
		CreatedAt *time.Time
	}

	now := time.Now()
	inner := &Inner{
		CreatedAt:        &now,
		Slice:            []string{"val1", "val2", "val3"},
		SliceStructs:     []*SliceStruct{{Name: "name1"}, {Name: "name2"}, {Name: "name3"}},
		SliceSlice:       [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}},
		SliceSliceStruct: [][]*SliceStruct{{{Name: "name1"}, {Name: "name2"}, {Name: "name3"}}, {{Name: "name4"}, {Name: "name5"}, {Name: "name6"}}, {{Name: "name7"}, {Name: "name8"}, {Name: "name9"}}},
		SliceMap:         []map[string]string{{"key1": "val1", "key2": "val2", "key3": "val3"}, {"key4": "val4", "key5": "val5", "key6": "val6"}},
		Map:              map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"},
		MapStructs:       map[string]*SliceStruct{"key1": {Name: "name1"}, "key2": {Name: "name2"}, "key3": {Name: "name3"}},
		MapMap:           map[string]map[string]string{"key1": {"key1-1": "val1"}, "key2": {"key2-1": "val2"}, "key3": {"key3-1": "val3"}},
		MapMapStruct:     map[string]map[string]*SliceStruct{"key1": {"key1-1": {Name: "name1"}}, "key2": {"key2-1": {Name: "name2"}}, "key3": {"key3-1": {Name: "name3"}}},
		MapSlice:         map[string][]string{"key1": {"1", "2", "3"}, "key2": {"4", "5", "6"}, "key3": {"7", "8", "9"}},
		MapInt:           map[int]string{1: "val1", 2: "val2", 3: "val3"},
		MapInt8:          map[int8]string{1: "val1", 2: "val2", 3: "val3"},
		MapInt16:         map[int16]string{1: "val1", 2: "val2", 3: "val3"},
		MapInt32:         map[int32]string{1: "val1", 2: "val2", 3: "val3"},
		MapInt64:         map[int64]string{1: "val1", 2: "val2", 3: "val3"},
		MapUint:          map[uint]string{1: "val1", 2: "val2", 3: "val3"},
		MapUint8:         map[uint8]string{1: "val1", 2: "val2", 3: "val3"},
		MapUint16:        map[uint16]string{1: "val1", 2: "val2", 3: "val3"},
		MapUint32:        map[uint32]string{1: "val1", 2: "val2", 3: "val3"},
		MapUint64:        map[uint64]string{1: "val1", 2: "val2", 3: "val3"},
		MapFloat32:       map[float32]string{1.01: "val1", 2.02: "val2", 3.03: "val3"},
		MapFloat64:       map[float64]string{1.01: "val1", 2.02: "val2", 3.03: "val3"},
		MapBool:          map[bool]string{true: "val1", false: "val2"},
	}

	test := &Test{
		Inner:     inner,
		CreatedAt: &now,
	}

	val := reflect.ValueOf(test)
	vd := New()
	v := &validate{
		v: vd,
	}

	current, kind, _, ok := v.getStructFieldOKInternal(val, "Inner.CreatedAt")
	Equal(t, ok, true)
	Equal(t, kind, reflect.Struct)
	tm, ok := current.Interface().(time.Time)
	Equal(t, ok, true)
	Equal(t, tm, now)

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.Slice[1]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	_, _, _, ok = v.getStructFieldOKInternal(val, "Inner.CrazyNonExistantField")
	Equal(t, ok, false)

	_, _, _, ok = v.getStructFieldOKInternal(val, "Inner.Slice[101]")
	Equal(t, ok, false)

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.Map[key3]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val3")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapMap[key2][key2-1]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapStructs[key2].Name")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "name2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapMapStruct[key3][key3-1].Name")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "name3")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.SliceSlice[2][0]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "7")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.SliceSliceStruct[2][1].Name")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "name8")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.SliceMap[1][key5]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val5")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapSlice[key3][2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "9")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapInt[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapInt8[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapInt16[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapInt32[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapInt64[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapUint[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapUint8[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapUint16[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapUint32[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapUint64[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapFloat32[3.03]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val3")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapFloat64[2.02]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val2")

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.MapBool[true]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.String)
	Equal(t, current.String(), "val1")

	inner = &Inner{
		CreatedAt:        &now,
		Slice:            []string{"val1", "val2", "val3"},
		SliceStructs:     []*SliceStruct{{Name: "name1"}, {Name: "name2"}, nil},
		SliceSlice:       [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}},
		SliceSliceStruct: [][]*SliceStruct{{{Name: "name1"}, {Name: "name2"}, {Name: "name3"}}, {{Name: "name4"}, {Name: "name5"}, {Name: "name6"}}, {{Name: "name7"}, {Name: "name8"}, {Name: "name9"}}},
		SliceMap:         []map[string]string{{"key1": "val1", "key2": "val2", "key3": "val3"}, {"key4": "val4", "key5": "val5", "key6": "val6"}},
		Map:              map[string]string{"key1": "val1", "key2": "val2", "key3": "val3"},
		MapStructs:       map[string]*SliceStruct{"key1": {Name: "name1"}, "key2": {Name: "name2"}, "key3": {Name: "name3"}},
		MapMap:           map[string]map[string]string{"key1": {"key1-1": "val1"}, "key2": {"key2-1": "val2"}, "key3": {"key3-1": "val3"}},
		MapMapStruct:     map[string]map[string]*SliceStruct{"key1": {"key1-1": {Name: "name1"}}, "key2": {"key2-1": {Name: "name2"}}, "key3": {"key3-1": {Name: "name3"}}},
		MapSlice:         map[string][]string{"key1": {"1", "2", "3"}, "key2": {"4", "5", "6"}, "key3": {"7", "8", "9"}},
	}

	test = &Test{
		Inner:     inner,
		CreatedAt: nil,
	}

	val = reflect.ValueOf(test)

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.SliceStructs[2]")
	Equal(t, ok, true)
	Equal(t, kind, reflect.Ptr)
	Equal(t, current.String(), "<*validator.SliceStruct Value>")
	Equal(t, current.IsNil(), true)

	current, kind, _, ok = v.getStructFieldOKInternal(val, "Inner.SliceStructs[2].Name")
	Equal(t, ok, false)
	Equal(t, kind, reflect.Ptr)
	Equal(t, current.String(), "<*validator.SliceStruct Value>")
	Equal(t, current.IsNil(), true)

	PanicMatches(t, func() { v.getStructFieldOKInternal(reflect.ValueOf(1), "crazyinput") }, "Invalid field namespace")
}

func TestExcludesValidation(t *testing.T) {
	tests := []struct {
		Value       string `validate:"excludes=@"`
		Tag         string
		ExpectedNil bool
	}{
		{Value: "abcd@!jfk", Tag: "excludes=@", ExpectedNil: false},
		{Value: "abcdq!jfk", Tag: "excludes=@", ExpectedNil: true},
	}

	validate := New()
	for i, s := range tests {
		errs := validate.Var(s.Value, s.Tag)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}

		errs = validate.Struct(s)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}
	}
}

func TestContainsRuneValidation(t *testing.T) {
	tests := []struct {
		Value       string `validate:"containsrune=☻"`
		Tag         string
		ExpectedNil bool
	}{
		{Value: "a☺b☻c☹d", Tag: "containsrune=☻", ExpectedNil: true},
		{Value: "abcd", Tag: "containsrune=☻", ExpectedNil: false},
	}

	validate := New()
	for i, s := range tests {
		errs := validate.Var(s.Value, s.Tag)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}

		errs = validate.Struct(s)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}
	}
}

func TestContainsAnyValidation(t *testing.T) {
	tests := []struct {
		Value       string `validate:"containsany=@!{}[]"`
		Tag         string
		ExpectedNil bool
	}{
		{Value: "abcd@!jfk", Tag: "containsany=@!{}[]", ExpectedNil: true},
		{Value: "abcdefg", Tag: "containsany=@!{}[]", ExpectedNil: false},
	}

	validate := New()
	for i, s := range tests {
		errs := validate.Var(s.Value, s.Tag)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}

		errs = validate.Struct(s)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}
	}
}

func TestContainsValidation(t *testing.T) {
	tests := []struct {
		Value       string `validate:"contains=@"`
		Tag         string
		ExpectedNil bool
	}{
		{Value: "abcd@!jfk", Tag: "contains=@", ExpectedNil: true},
		{Value: "abcdq!jfk", Tag: "contains=@", ExpectedNil: false},
	}

	validate := New()
	for i, s := range tests {
		errs := validate.Var(s.Value, s.Tag)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}

		errs = validate.Struct(s)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}
	}
}

func TestBase64URLValidation(t *testing.T) {
	validate := New()
	testCases := []struct {
		decoded, encoded string
		success          bool
	}{
		// empty string, although a valid base64 string, should fail
		{"", "", false},
		// invalid length
		{"", "a", false},
		// base64 with padding
		{"f", "Zg==", true},
		{"fo", "Zm8=", true},
		// base64 without padding
		{"foo", "Zm9v", true},
		{"", "Zg", false},
		{"", "Zm8", false},
		// base64 URL safe encoding with invalid, special characters '+' and '/'
		{"\x14\xfb\x9c\x03\xd9\x7e", "FPucA9l+", false},
		{"\x14\xfb\x9c\x03\xf9\x73", "FPucA/lz", false},
		// base64 URL safe encoding with valid, special characters '-' and '_'
		{"\x14\xfb\x9c\x03\xd9\x7e", "FPucA9l-", true},
		{"\x14\xfb\x9c\x03\xf9\x73", "FPucA_lz", true},
		// non base64 characters
		{"", "@mc=", false},
		{"", "Zm 9", false},
	}

	for _, tc := range testCases {
		err := validate.Var(tc.encoded, "base64url")
		if tc.success {
			Equal(t, err, nil)
			// make sure encoded value is decoded back to the expected value
			d, innerErr := base64.URLEncoding.DecodeString(tc.encoded)
			Equal(t, innerErr, nil)
			Equal(t, tc.decoded, string(d))
		} else {
			NotEqual(t, err, nil)
			if len(tc.encoded) > 0 {
				// make sure that indeed the encoded value was faulty
				_, err := base64.URLEncoding.DecodeString(tc.encoded)
				NotEqual(t, err, nil)
			}
		}
	}
}

func TestBase64RawURLValidation(t *testing.T) {
	validate := New()
	testCases := []struct {
		decoded, encoded string
		success          bool
	}{
		// empty string, although a valid base64 string, should fail
		{"", "", false},
		// invalid length
		{"", "a", false},
		// base64 with padding should fail
		{"f", "Zg==", false},
		{"fo", "Zm8=", false},
		// base64 without padding
		{"foo", "Zm9v", true},
		{"hello", "aGVsbG8", true},
		{"", "aGVsb", false},
		// // base64 URL safe encoding with invalid, special characters '+' and '/'
		{"\x14\xfb\x9c\x03\xd9\x7e", "FPucA9l+", false},
		{"\x14\xfb\x9c\x03\xf9\x73", "FPucA/lz", false},
		// // base64 URL safe encoding with valid, special characters '-' and '_'
		{"\x14\xfb\x9c\x03\xd9\x7e", "FPucA9l-", true},
		{"\x14\xfb\x9c\x03\xf9\x73", "FPucA_lz", true},
		// non base64 characters
		{"", "@mc=", false},
		{"", "Zm 9", false},
	}

	for _, tc := range testCases {
		err := validate.Var(tc.encoded, "base64rawurl")
		if tc.success {
			Equal(t, err, nil)
			// make sure encoded value is decoded back to the expected value
			d, innerErr := base64.RawURLEncoding.DecodeString(tc.encoded)
			Equal(t, innerErr, nil)
			Equal(t, tc.decoded, string(d))
		} else {
			NotEqual(t, err, nil)
			if len(tc.encoded) > 0 {
				// make sure that indeed the encoded value was faulty
				_, err := base64.RawURLEncoding.DecodeString(tc.encoded)
				NotEqual(t, err, nil)
			}
		}
	}
}

func TestFileValidation(t *testing.T) {
	validate := New()
	tests := []struct {
		title    string
		param    string
		expected bool
	}{
		{"empty path", "", false},
		{"regular file", filepath.Join("testdata", "a.go"), true},
		{"missing file", filepath.Join("testdata", "no.go"), false},
		{"directory, not a file", "testdata", false},
	}

	for _, test := range tests {
		errs := validate.Var(test.param, "file")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		}
	}

	PanicMatches(t, func() {
		_ = validate.Var(6, "file")
	}, "Bad field type int")
}

func TestImageValidation(t *testing.T) {
	validate := New()
	tmpDir := t.TempDir()
	paths := map[string]string{
		"empty":     "",
		"directory": "testdata",
		"missing":   filepath.Join(tmpDir, "none.png"),
		"png":       filepath.Join(tmpDir, "image.png"),
		"jpeg":      filepath.Join(tmpDir, "image.jpg"),
		"mp3":       filepath.Join(tmpDir, "music.mp3"),
	}

	tests := []struct {
		title        string
		param        string
		expected     bool
		createImage  func()
		destroyImage func()
	}{
		{
			"empty path",
			paths["empty"], false,
			func() {},
			func() {},
		},
		{
			"directory, not a file",
			paths["directory"],
			false,
			func() {},
			func() {},
		},
		{
			"missing file",
			paths["missing"],
			false,
			func() {},
			func() {},
		},
		{
			"valid png",
			paths["png"],
			true,
			func() {
				img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{10, 10}})
				f, err := os.Create(paths["png"])
				Equal(t, err, nil)
				defer func() {
					_ = f.Close()
				}()

				err = png.Encode(f, img)
				Equal(t, err, nil)
			},
			func() {
				err := os.Remove(paths["png"])
				Equal(t, err, nil)
			},
		},
		{
			"valid jpeg",
			paths["jpeg"],
			true,
			func() {
				var opt jpeg.Options
				img := image.NewGray(image.Rect(0, 0, 10, 10))
				f, err := os.Create(paths["jpeg"])
				Equal(t, err, nil)
				defer func() {
					_ = f.Close()
				}()

				err = jpeg.Encode(f, img, &opt)
				Equal(t, err, nil)
			},
			func() {
				err := os.Remove(paths["jpeg"])
				Equal(t, err, nil)
			},
		},
		{
			"valid mp3",
			paths["mp3"],
			false,
			func() {},
			func() {},
		},
	}

	for _, test := range tests {
		test.createImage()
		errs := validate.Var(test.param, "image")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		}
		test.destroyImage()
	}

	PanicMatches(t, func() {
		_ = validate.Var(6, "image")
	}, "Bad field type int")
}

func TestFilePathValidation(t *testing.T) {
	validate := New()
	tests := []struct {
		title    string
		param    string
		expected bool
	}{
		{"empty filepath", "", false},
		{"valid filepath", filepath.Join("testdata", "a.go"), true},
		{"invalid filepath", filepath.Join("testdata", "no\000.go"), false},
		{"directory, not a filepath", "testdata" + string(os.PathSeparator), false},
		{"directory", "testdata", false},
	}

	for _, test := range tests {
		errs := validate.Var(test.param, "filepath")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		}
	}

	PanicMatches(t, func() {
		_ = validate.Var(6, "filepath")
	}, "Bad field type int")
}

func TestDirValidation(t *testing.T) {
	validate := New()
	tests := []struct {
		title    string
		param    string
		expected bool
	}{
		{"existing dir", "testdata", true},
		{"existing self dir", ".", true},
		{"existing parent dir", "..", true},
		{"empty dir", "", false},
		{"missing dir", "non_existing_testdata", false},
		{"a file not a directory", filepath.Join("testdata", "a.go"), false},
	}

	for _, test := range tests {
		errs := validate.Var(test.param, "dir")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		}
	}

	PanicMatches(t, func() {
		_ = validate.Var(2, "dir")
	}, "Bad field type int")
}

func TestDirPathValidation(t *testing.T) {
	validate := New()
	tests := []struct {
		title    string
		param    string
		expected bool
	}{
		{"empty dirpath", "", false},
		{"valid dirpath - exists", "testdata", true},
		{"valid dirpath - explicit", "testdatanoexist" + string(os.PathSeparator), true},
		{"invalid dirpath", "testdata\000" + string(os.PathSeparator), false},
		{"file, not a dirpath", filepath.Join("testdata", "a.go"), false},
	}

	for _, test := range tests {
		errs := validate.Var(test.param, "dirpath")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Test: '%s' failed Error: %s", test.title, errs)
			}
		}
	}

	PanicMatches(t, func() {
		_ = validate.Var(6, "filepath")
	}, "Bad field type int")
}

func TestStartsWithValidation(t *testing.T) {
	tests := []struct {
		Value       string `validate:"startswith=(/^ヮ^)/*:・ﾟ✧"`
		Tag         string
		ExpectedNil bool
	}{
		{Value: "(/^ヮ^)/*:・ﾟ✧ glitter", Tag: "startswith=(/^ヮ^)/*:・ﾟ✧", ExpectedNil: true},
		{Value: "abcd", Tag: "startswith=(/^ヮ^)/*:・ﾟ✧", ExpectedNil: false},
	}

	validate := New()
	for i, s := range tests {
		errs := validate.Var(s.Value, s.Tag)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}

		errs = validate.Struct(s)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}
	}
}

func TestEndsWithValidation(t *testing.T) {
	tests := []struct {
		Value       string `validate:"endswith=(/^ヮ^)/*:・ﾟ✧"`
		Tag         string
		ExpectedNil bool
	}{
		{Value: "glitter (/^ヮ^)/*:・ﾟ✧", Tag: "endswith=(/^ヮ^)/*:・ﾟ✧", ExpectedNil: true},
		{Value: "(/^ヮ^)/*:・ﾟ✧ glitter", Tag: "endswith=(/^ヮ^)/*:・ﾟ✧", ExpectedNil: false},
	}

	validate := New()
	for i, s := range tests {
		errs := validate.Var(s.Value, s.Tag)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}

		errs = validate.Struct(s)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}
	}
}

func TestLookup(t *testing.T) {
	type Lookup struct {
		FieldA *string `json:"fieldA,omitempty" validate:"required_without=FieldB"`
		FieldB *string `json:"fieldB,omitempty" validate:"required_without=FieldA"`
	}

	fieldAValue := "1232"
	lookup := Lookup{
		FieldA: &fieldAValue,
		FieldB: nil,
	}
	Equal(t, New().Struct(lookup), nil)
}

func TestAbilityToValidateNils(t *testing.T) {
	type TestStruct struct {
		Test *string `validate:"nil"`
	}

	ts := TestStruct{}
	val := New()
	fn := func(fl FieldLevel) bool {
		return fl.Field().Kind() == reflect.Ptr && fl.Field().IsNil()
	}

	err := val.RegisterValidation("nil", fn, true)
	Equal(t, err, nil)

	errs := val.Struct(ts)
	Equal(t, errs, nil)

	str := "string"
	ts.Test = &str

	errs = val.Struct(ts)
	NotEqual(t, errs, nil)
}

func TestRequiredWithoutPointers(t *testing.T) {
	type Lookup struct {
		FieldA *bool `json:"fieldA,omitempty" validate:"required_without=FieldB"`
		FieldB *bool `json:"fieldB,omitempty" validate:"required_without=FieldA"`
	}

	b := true
	lookup := Lookup{
		FieldA: &b,
		FieldB: nil,
	}

	val := New()
	errs := val.Struct(lookup)
	Equal(t, errs, nil)

	b = false
	lookup = Lookup{
		FieldA: &b,
		FieldB: nil,
	}
	errs = val.Struct(lookup)
	Equal(t, errs, nil)
}

func TestRequiredWithoutAllPointers(t *testing.T) {
	type Lookup struct {
		FieldA *bool `json:"fieldA,omitempty" validate:"required_without_all=FieldB"`
		FieldB *bool `json:"fieldB,omitempty" validate:"required_without_all=FieldA"`
	}

	b := true
	lookup := Lookup{
		FieldA: &b,
		FieldB: nil,
	}

	val := New()
	errs := val.Struct(lookup)
	Equal(t, errs, nil)

	b = false
	lookup = Lookup{
		FieldA: &b,
		FieldB: nil,
	}
	errs = val.Struct(lookup)
	Equal(t, errs, nil)
}

func TestGetTag(t *testing.T) {
	var tag string

	type Test struct {
		String string `validate:"mytag"`
	}

	val := New()
	_ = val.RegisterValidation("mytag", func(fl FieldLevel) bool {
		tag = fl.GetTag()
		return true
	})

	var test Test
	errs := val.Struct(test)
	Equal(t, errs, nil)
	Equal(t, tag, "mytag")
}

func TestIsIso3166Alpha2Validation(t *testing.T) {
	tests := []struct {
		value    string `validate:"iso3166_1_alpha2"`
		expected bool
	}{
		{"PL", true},
		{"POL", false},
		{"AA", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso3166_1_alpha2")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha2 failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha2 failed Error: %s", i, errs)
			}
		}
	}
}

func TestIsIso3166Alpha2EUValidation(t *testing.T) {
	tests := []struct {
		value    string `validate:"iso3166_1_alpha2_eu"`
		expected bool
	}{
		{"SE", true},
		{"UK", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso3166_1_alpha2_eu")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha2_eu failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha2_eu failed Error: %s", i, errs)
			}
		}
	}
}

func TestIsIso31662Validation(t *testing.T) {
	tests := []struct {
		value    string `validate:"iso3166_2"`
		expected bool
	}{
		{"US-FL", true},
		{"US-F", false},
		{"US", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso3166_2")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_2 failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_2 failed Error: %s", i, errs)
			}
		}
	}
}

func TestIsIso3166Alpha3Validation(t *testing.T) {
	tests := []struct {
		value    string `validate:"iso3166_1_alpha3"`
		expected bool
	}{
		{"POL", true},
		{"PL", false},
		{"AAA", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso3166_1_alpha3")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha3 failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha3 failed Error: %s", i, errs)
			}
		}
	}
}

func TestIsIso3166Alpha3EUValidation(t *testing.T) {
	tests := []struct {
		value    string `validate:"iso3166_1_alpha3_eu"`
		expected bool
	}{
		{"POL", true},
		{"SWE", true},
		{"UNK", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso3166_1_alpha3_eu")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha3_eu failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha3_eu failed Error: %s", i, errs)
			}
		}
	}
}

func TestIsIso3166AlphaNumericValidation(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected bool
	}{
		{248, true},
		{"248", true},
		{0, false},
		{1, false},
		{"1", false},
		{"invalid_int", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso3166_1_alpha_numeric")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha_numeric failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha_numeric failed Error: %s", i, errs)
			}
		}
	}

	PanicMatches(t, func() {
		_ = validate.Var([]string{"1"}, "iso3166_1_alpha_numeric")
	}, "Bad field type []string")
}

func TestIsIso3166AlphaNumericEUValidation(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected bool
	}{
		{752, true}, // Sweden
		{"752", true},
		{826, false}, // UK
		{"826", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso3166_1_alpha_numeric_eu")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha_numeric_eu failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso3166_1_alpha_numeric_eu failed Error: %s", i, errs)
			}
		}
	}

	PanicMatches(t, func() {
		_ = validate.Var([]string{"1"}, "iso3166_1_alpha_numeric_eu")
	}, "Bad field type []string")
}

func TestCountryCodeValidation(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected bool
	}{
		{248, true},
		{0, false},
		{1, false},
		{"POL", true},
		{"NO", true},
		{"248", true},
		{"1", false},
		{"0", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "country_code")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d country_code failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d country_code failed Error: %s", i, errs)
			}
		}
	}
}

func TestEUCountryCodeValidation(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected bool
	}{
		{724, true},
		{0, false},
		{1, false},
		{"POL", true},
		{"NO", false},
		{"724", true},
		{"1", false},
		{"0", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "eu_country_code")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d eu_country_code failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d eu_country_code failed Error: %s", i, errs)
			}
		}
	}
}

func TestIsIso4217Validation(t *testing.T) {
	tests := []struct {
		value    string `validate:"iso4217"`
		expected bool
	}{
		{"TRY", true},
		{"EUR", true},
		{"USA", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso4217")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso4217 failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso4217 failed Error: %s", i, errs)
			}
		}
	}
}

func TestIsIso4217NumericValidation(t *testing.T) {
	tests := []struct {
		value    int `validate:"iso4217_numeric"`
		expected bool
	}{
		{8, true},
		{12, true},
		{13, false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "iso4217_numeric")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso4217 failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d iso4217 failed Error: %s", i, errs)
			}
		}
	}
}

func TestPostCodeByIso3166Alpha2(t *testing.T) {
	tests := map[string][]struct {
		value    string
		expected bool
	}{
		"VN": {
			{"ABC", false},
			{"700000", true},
			{"A1", false},
		},
		"GB": {
			{"EC1A 1BB", true},
			{"CF10 1B1H", false},
		},
		"VI": {
			{"00803", true},
			{"1234567", false},
		},
		"LC": {
			// not support regexp for post code
			{"123456", false},
		},
		"XX": {
			// not support country
			{"123456", false},
		},
	}

	validate := New()
	for cc, ccTests := range tests {
		for i, test := range ccTests {
			errs := validate.Var(test.value, fmt.Sprintf("postcode_iso3166_alpha2=%s", cc))
			if test.expected {
				if !IsEqual(errs, nil) {
					t.Fatalf("Index: %d postcode_iso3166_alpha2=%s failed Error: %s", i, cc, errs)
				}
			} else {
				if IsEqual(errs, nil) {
					t.Fatalf("Index: %d postcode_iso3166_alpha2=%s failed Error: %s", i, cc, errs)
				}
			}
		}
	}
}

func TestPostCodeByIso3166Alpha2Field(t *testing.T) {
	tests := []struct {
		Value       string `validate:"postcode_iso3166_alpha2_field=CountryCode"`
		CountryCode interface{}
		expected    bool
	}{
		{"ABC", "VN", false},
		{"700000", "VN", true},
		{"A1", "VN", false},
		{"EC1A 1BB", "GB", true},
		{"CF10 1B1H", "GB", false},
		{"00803", "VI", true},
		{"1234567", "VI", false},
		{"123456", "LC", false}, // not support regexp for post code
		{"123456", "XX", false}, // not support country
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Struct(test)
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d postcode_iso3166_alpha2_field=CountryCode failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d postcode_iso3166_alpha2_field=CountryCode failed Error: %s", i, errs)
			}
		}
	}
}

func TestPostCodeByIso3166Alpha2Field_WrongField(t *testing.T) {
	type test struct {
		Value        string `validate:"postcode_iso3166_alpha2_field=CountryCode"`
		CountryCode1 interface{}
		expected     bool
	}

	errs := New().Struct(test{"ABC", "VN", false})
	NotEqual(t, nil, errs)
}

func TestPostCodeByIso3166Alpha2Field_MissingParam(t *testing.T) {
	type test struct {
		Value        string `validate:"postcode_iso3166_alpha2_field="`
		CountryCode1 interface{}
		expected     bool
	}

	errs := New().Struct(test{"ABC", "VN", false})
	NotEqual(t, nil, errs)
}

func TestPostCodeByIso3166Alpha2Field_InvalidKind(t *testing.T) {
	type test struct {
		Value       string `validate:"postcode_iso3166_alpha2_field=CountryCode"`
		CountryCode interface{}
		expected    bool
	}
	defer func() { _ = recover() }()

	_ = New().Struct(test{"ABC", 123, false})
	t.Errorf("Didn't panic as expected")
}

func TestValidate_ValidateMapCtx(t *testing.T) {
	type args struct {
		data  map[string]interface{}
		rules map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test nested map in slice",
			args: args{
				data: map[string]interface{}{
					"Test_A": map[string]interface{}{
						"Test_B": "Test_B",
						"Test_C": []map[string]interface{}{
							{
								"Test_D": "Test_D",
							},
						},
						"Test_E": map[string]interface{}{
							"Test_F": "Test_F",
						},
					},
				},
				rules: map[string]interface{}{
					"Test_A": map[string]interface{}{
						"Test_B": "min=2",
						"Test_C": map[string]interface{}{
							"Test_D": "min=2",
						},
						"Test_E": map[string]interface{}{
							"Test_F": "min=2",
						},
					},
				},
			},
			want: 0,
		},

		{
			name: "test nested map error",
			args: args{
				data: map[string]interface{}{
					"Test_A": map[string]interface{}{
						"Test_B": "Test_B",
						"Test_C": []interface{}{"Test_D"},
						"Test_E": map[string]interface{}{
							"Test_F": "Test_F",
						},
						"Test_G": "Test_G",
						"Test_I": []map[string]interface{}{
							{
								"Test_J": "Test_J",
							},
						},
					},
				},
				rules: map[string]interface{}{
					"Test_A": map[string]interface{}{
						"Test_B": "min=2",
						"Test_C": map[string]interface{}{
							"Test_D": "min=2",
						},
						"Test_E": map[string]interface{}{
							"Test_F": "min=100",
						},
						"Test_G": map[string]interface{}{
							"Test_H": "min=2",
						},
						"Test_I": map[string]interface{}{
							"Test_J": "min=100",
						},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validate := New()
			if got := validate.ValidateMapCtx(context.Background(), tt.args.data, tt.args.rules); len(got) != tt.want {
				t.Errorf("ValidateMapCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEINStringValidation(t *testing.T) {
	tests := []struct {
		value    string `validate:"ein"`
		expected bool
	}{
		{"01-2564282", true},
		{"25-4573894", true},
		{"63-236", false},
		{"3-5738294", false},
		{"4235-48", false},
		{"0.-47829", false},
		{"23-", false},
	}
	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, "ein")
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d ein failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d ein failed Error: %s", i, errs)
			}
		}
	}
}

func TestInterfaceErrValidation(t *testing.T) {
	var v2 interface{} = 1
	var v1 = v2
	validate := New()
	errs := validate.Var(v1, "len=1")
	Equal(t, errs, nil)

	errs = validate.Var(v2, "len=1")
	Equal(t, errs, nil)

	type ExternalCMD struct {
		Userid string      `json:"userid"`
		Action uint32      `json:"action"`
		Data   interface{} `json:"data,omitempty" validate:"required"`
	}

	s := &ExternalCMD{
		Userid: "123456",
		Action: 10000,
		// Data:   1,
	}

	errs = validate.Struct(s)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "ExternalCMD.Data", "ExternalCMD.Data", "Data", "Data", "required")

	type ExternalCMD2 struct {
		Userid string      `json:"userid"`
		Action uint32      `json:"action"`
		Data   interface{} `json:"data,omitempty" validate:"len=1"`
	}

	s2 := &ExternalCMD2{
		Userid: "123456",
		Action: 10000,
		// Data:   1,
	}

	errs = validate.Struct(s2)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "ExternalCMD2.Data", "ExternalCMD2.Data", "Data", "Data", "len")

	s3 := &ExternalCMD2{
		Userid: "123456",
		Action: 10000,
		Data:   2,
	}

	errs = validate.Struct(s3)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "ExternalCMD2.Data", "ExternalCMD2.Data", "Data", "Data", "len")

	type Inner struct {
		Name string `validate:"required"`
	}

	inner := &Inner{
		Name: "",
	}

	s4 := &ExternalCMD{
		Userid: "123456",
		Action: 10000,
		Data:   inner,
	}

	errs = validate.Struct(s4)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "ExternalCMD.Data.Name", "ExternalCMD.Data.Name", "Name", "Name", "required")

	type TestMapStructPtr struct {
		Errs map[int]interface{} `validate:"gt=0,dive,required"`
	}

	mip := map[int]interface{}{0: &Inner{"ok"}, 3: nil, 4: &Inner{"ok"}}

	msp := &TestMapStructPtr{
		Errs: mip,
	}

	errs = validate.Struct(msp)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "TestMapStructPtr.Errs[3]", "TestMapStructPtr.Errs[3]", "Errs[3]", "Errs[3]", "required")

	type TestMultiDimensionalStructs struct {
		Errs [][]interface{} `validate:"gt=0,dive,dive"`
	}

	var errStructArray [][]interface{}
	errStructArray = append(errStructArray, []interface{}{&Inner{"ok"}, &Inner{""}, &Inner{""}})
	errStructArray = append(errStructArray, []interface{}{&Inner{"ok"}, &Inner{""}, &Inner{""}})
	tms := &TestMultiDimensionalStructs{
		Errs: errStructArray,
	}

	errs = validate.Struct(tms)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 4)
	AssertError(t, errs, "TestMultiDimensionalStructs.Errs[0][1].Name", "TestMultiDimensionalStructs.Errs[0][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructs.Errs[0][2].Name", "TestMultiDimensionalStructs.Errs[0][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructs.Errs[1][1].Name", "TestMultiDimensionalStructs.Errs[1][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructs.Errs[1][2].Name", "TestMultiDimensionalStructs.Errs[1][2].Name", "Name", "Name", "required")

	type TestMultiDimensionalStructsPtr2 struct {
		Errs [][]*Inner `validate:"gt=0,dive,dive,required"`
	}

	var errStructPtr2Array [][]*Inner
	errStructPtr2Array = append(errStructPtr2Array, []*Inner{{"ok"}, {""}, {""}})
	errStructPtr2Array = append(errStructPtr2Array, []*Inner{{"ok"}, {""}, {""}})
	errStructPtr2Array = append(errStructPtr2Array, []*Inner{{"ok"}, {""}, nil})
	tmsp2 := &TestMultiDimensionalStructsPtr2{
		Errs: errStructPtr2Array,
	}

	errs = validate.Struct(tmsp2)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 6)
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[0][1].Name", "TestMultiDimensionalStructsPtr2.Errs[0][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[0][2].Name", "TestMultiDimensionalStructsPtr2.Errs[0][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[1][1].Name", "TestMultiDimensionalStructsPtr2.Errs[1][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[1][2].Name", "TestMultiDimensionalStructsPtr2.Errs[1][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[2][1].Name", "TestMultiDimensionalStructsPtr2.Errs[2][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[2][2]", "TestMultiDimensionalStructsPtr2.Errs[2][2]", "Errs[2][2]", "Errs[2][2]", "required")

	m := map[int]interface{}{0: "ok", 3: "", 4: "ok"}
	errs = validate.Var(m, "len=3,dive,len=2")
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "[3]", "[3]", "[3]", "[3]", "len")

	errs = validate.Var(m, "len=2,dive,required")
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "", "", "", "", "len")

	arr := []interface{}{"ok", "", "ok"}
	errs = validate.Var(arr, "len=3,dive,len=2")
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "[1]", "[1]", "[1]", "[1]", "len")

	errs = validate.Var(arr, "len=2,dive,required")
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "", "", "", "", "len")

	type MyStruct struct {
		A, B string
		C    interface{}
	}

	var a MyStruct
	a.A = "value"
	a.C = "nu"
	errs = validate.Struct(a)
	Equal(t, errs, nil)
}

func TestAnonymousSameStructDifferentTags(t *testing.T) {
	validate := New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "-" {
			return name
		}
		return ""
	})

	type Test struct {
		A interface{}
	}

	tst := &Test{
		A: struct {
			A string `validate:"required"`
		}{
			A: "",
		},
	}

	err := validate.Struct(tst)
	NotEqual(t, err, nil)

	errs := err.(ValidationErrors)
	Equal(t, len(errs), 1)
	AssertError(t, errs, "Test.A.A", "Test.A.A", "A", "A", "required")

	tst = &Test{
		A: struct {
			A string `validate:"omitempty,required"`
		}{
			A: "",
		},
	}
	err = validate.Struct(tst)
	Equal(t, err, nil)
}

func TestCrossStructLteFieldValidation(t *testing.T) {
	var errs error
	validate := New()
	type Inner struct {
		CreatedAt *time.Time
		String    string
		Int       int
		Uint      uint
		Float     float64
		Array     []string
	}

	type Test struct {
		Inner     *Inner
		CreatedAt *time.Time `validate:"ltecsfield=Inner.CreatedAt"`
		String    string     `validate:"ltecsfield=Inner.String"`
		Int       int        `validate:"ltecsfield=Inner.Int"`
		Uint      uint       `validate:"ltecsfield=Inner.Uint"`
		Float     float64    `validate:"ltecsfield=Inner.Float"`
		Array     []string   `validate:"ltecsfield=Inner.Array"`
	}

	now := time.Now().UTC()
	then := now.Add(time.Hour * 5)
	inner := &Inner{
		CreatedAt: &then,
		String:    "abcd",
		Int:       13,
		Uint:      13,
		Float:     1.13,
		Array:     []string{"val1", "val2"},
	}

	test := &Test{
		Inner:     inner,
		CreatedAt: &now,
		String:    "abc",
		Int:       12,
		Uint:      12,
		Float:     1.12,
		Array:     []string{"val1"},
	}

	errs = validate.Struct(test)
	Equal(t, errs, nil)

	test.CreatedAt = &then
	test.String = "abcd"
	test.Int = 13
	test.Uint = 13
	test.Float = 1.13
	test.Array = []string{"val1", "val2"}

	errs = validate.Struct(test)
	Equal(t, errs, nil)

	after := now.Add(time.Hour * 10)

	test.CreatedAt = &after
	test.String = "abce"
	test.Int = 14
	test.Uint = 14
	test.Float = 1.14
	test.Array = []string{"val1", "val2", "val3"}

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "ltecsfield")
	AssertError(t, errs, "Test.String", "Test.String", "String", "String", "ltecsfield")
	AssertError(t, errs, "Test.Int", "Test.Int", "Int", "Int", "ltecsfield")
	AssertError(t, errs, "Test.Uint", "Test.Uint", "Uint", "Uint", "ltecsfield")
	AssertError(t, errs, "Test.Float", "Test.Float", "Float", "Float", "ltecsfield")
	AssertError(t, errs, "Test.Array", "Test.Array", "Array", "Array", "ltecsfield")

	errs = validate.VarWithValueCtx(context.Background(), 1, "", "ltecsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltecsfield")

	// this test is for the WARNING about unforeseen validation issues.
	errs = validate.VarWithValue(test, now, "ltecsfield")
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 6)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "ltecsfield")
	AssertError(t, errs, "Test.String", "Test.String", "String", "String", "ltecsfield")
	AssertError(t, errs, "Test.Int", "Test.Int", "Int", "Int", "ltecsfield")
	AssertError(t, errs, "Test.Uint", "Test.Uint", "Uint", "Uint", "ltecsfield")
	AssertError(t, errs, "Test.Float", "Test.Float", "Float", "Float", "ltecsfield")
	AssertError(t, errs, "Test.Array", "Test.Array", "Array", "Array", "ltecsfield")

	type Other struct {
		Value string
	}

	type Test2 struct {
		Value Other
		Time  time.Time `validate:"ltecsfield=Value"`
	}

	tst := Test2{
		Value: Other{Value: "StringVal"},
		Time:  then,
	}
	errs = validate.Struct(tst)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test2.Time", "Test2.Time", "Time", "Time", "ltecsfield")

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "ltecsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "ltecsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "ltecsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltecsfield")

	errs = validate.VarWithValue(time.Duration(0), -time.Minute, "omitempty,ltecsfield")
	Equal(t, errs, nil)

	// -- Validations for a struct and an inner struct with time.Duration type fields.

	type TimeDurationInner struct {
		Duration time.Duration
	}

	var timeDurationInner *TimeDurationInner

	type TimeDurationTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"ltecsfield=Inner.Duration"`
	}

	var timeDurationTest *TimeDurationTest

	timeDurationInner = &TimeDurationInner{time.Hour + time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour - time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "ltecsfield")

	type TimeDurationOmitemptyTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"omitempty,ltecsfield=Inner.Duration"`
	}

	var timeDurationOmitemptyTest *TimeDurationOmitemptyTest
	timeDurationInner = &TimeDurationInner{-time.Minute}
	timeDurationOmitemptyTest = &TimeDurationOmitemptyTest{timeDurationInner, time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestCrossStructLtFieldValidation(t *testing.T) {
	var errs error
	validate := New()
	type Inner struct {
		CreatedAt *time.Time
		String    string
		Int       int
		Uint      uint
		Float     float64
		Array     []string
	}

	type Test struct {
		Inner     *Inner
		CreatedAt *time.Time `validate:"ltcsfield=Inner.CreatedAt"`
		String    string     `validate:"ltcsfield=Inner.String"`
		Int       int        `validate:"ltcsfield=Inner.Int"`
		Uint      uint       `validate:"ltcsfield=Inner.Uint"`
		Float     float64    `validate:"ltcsfield=Inner.Float"`
		Array     []string   `validate:"ltcsfield=Inner.Array"`
	}

	now := time.Now().UTC()
	then := now.Add(time.Hour * 5)
	inner := &Inner{
		CreatedAt: &then,
		String:    "abcd",
		Int:       13,
		Uint:      13,
		Float:     1.13,
		Array:     []string{"val1", "val2"},
	}

	test := &Test{
		Inner:     inner,
		CreatedAt: &now,
		String:    "abc",
		Int:       12,
		Uint:      12,
		Float:     1.12,
		Array:     []string{"val1"},
	}

	errs = validate.Struct(test)
	Equal(t, errs, nil)

	test.CreatedAt = &then
	test.String = "abcd"
	test.Int = 13
	test.Uint = 13
	test.Float = 1.13
	test.Array = []string{"val1", "val2"}

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "ltcsfield")
	AssertError(t, errs, "Test.String", "Test.String", "String", "String", "ltcsfield")
	AssertError(t, errs, "Test.Int", "Test.Int", "Int", "Int", "ltcsfield")
	AssertError(t, errs, "Test.Uint", "Test.Uint", "Uint", "Uint", "ltcsfield")
	AssertError(t, errs, "Test.Float", "Test.Float", "Float", "Float", "ltcsfield")
	AssertError(t, errs, "Test.Array", "Test.Array", "Array", "Array", "ltcsfield")

	errs = validate.VarWithValue(1, "", "ltcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltcsfield")

	// this test is for the WARNING about unforeseen validation issues.
	errs = validate.VarWithValue(test, now, "ltcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "ltcsfield")
	AssertError(t, errs, "Test.String", "Test.String", "String", "String", "ltcsfield")
	AssertError(t, errs, "Test.Int", "Test.Int", "Int", "Int", "ltcsfield")
	AssertError(t, errs, "Test.Uint", "Test.Uint", "Uint", "Uint", "ltcsfield")
	AssertError(t, errs, "Test.Float", "Test.Float", "Float", "Float", "ltcsfield")
	AssertError(t, errs, "Test.Array", "Test.Array", "Array", "Array", "ltcsfield")

	type Other struct {
		Value string
	}

	type Test2 struct {
		Value Other
		Time  time.Time `validate:"ltcsfield=Value"`
	}

	tst := Test2{
		Value: Other{Value: "StringVal"},
		Time:  then,
	}
	errs = validate.Struct(tst)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test2.Time", "Test2.Time", "Time", "Time", "ltcsfield")

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "ltcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "ltcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltcsfield")

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "ltcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltcsfield")

	errs = validate.VarWithValue(time.Duration(0), -time.Minute, "omitempty,ltcsfield")
	Equal(t, errs, nil)

	// -- Validations for a struct and an inner struct with time.Duration type fields.

	type TimeDurationInner struct {
		Duration time.Duration
	}

	var timeDurationInner *TimeDurationInner

	type TimeDurationTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"ltcsfield=Inner.Duration"`
	}

	var timeDurationTest *TimeDurationTest

	timeDurationInner = &TimeDurationInner{time.Hour + time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "ltcsfield")

	timeDurationInner = &TimeDurationInner{time.Hour - time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "ltcsfield")

	type TimeDurationOmitemptyTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"omitempty,ltcsfield=Inner.Duration"`
	}

	var timeDurationOmitemptyTest *TimeDurationOmitemptyTest
	timeDurationInner = &TimeDurationInner{-time.Minute}
	timeDurationOmitemptyTest = &TimeDurationOmitemptyTest{timeDurationInner, time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestCrossStructGteFieldValidation(t *testing.T) {
	var errs error
	validate := New()
	type Inner struct {
		CreatedAt *time.Time
		String    string
		Int       int
		Uint      uint
		Float     float64
		Array     []string
	}

	type Test struct {
		Inner     *Inner
		CreatedAt *time.Time `validate:"gtecsfield=Inner.CreatedAt"`
		String    string     `validate:"gtecsfield=Inner.String"`
		Int       int        `validate:"gtecsfield=Inner.Int"`
		Uint      uint       `validate:"gtecsfield=Inner.Uint"`
		Float     float64    `validate:"gtecsfield=Inner.Float"`
		Array     []string   `validate:"gtecsfield=Inner.Array"`
	}

	now := time.Now().UTC()
	then := now.Add(time.Hour * -5)
	inner := &Inner{
		CreatedAt: &then,
		String:    "abcd",
		Int:       13,
		Uint:      13,
		Float:     1.13,
		Array:     []string{"val1", "val2"},
	}

	test := &Test{
		Inner:     inner,
		CreatedAt: &now,
		String:    "abcde",
		Int:       14,
		Uint:      14,
		Float:     1.14,
		Array:     []string{"val1", "val2", "val3"},
	}

	errs = validate.Struct(test)
	Equal(t, errs, nil)

	test.CreatedAt = &then
	test.String = "abcd"
	test.Int = 13
	test.Uint = 13
	test.Float = 1.13
	test.Array = []string{"val1", "val2"}

	errs = validate.Struct(test)
	Equal(t, errs, nil)

	before := now.Add(time.Hour * -10)

	test.CreatedAt = &before
	test.String = "abc"
	test.Int = 12
	test.Uint = 12
	test.Float = 1.12
	test.Array = []string{"val1"}

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "gtecsfield")
	AssertError(t, errs, "Test.String", "Test.String", "String", "String", "gtecsfield")
	AssertError(t, errs, "Test.Int", "Test.Int", "Int", "Int", "gtecsfield")
	AssertError(t, errs, "Test.Uint", "Test.Uint", "Uint", "Uint", "gtecsfield")
	AssertError(t, errs, "Test.Float", "Test.Float", "Float", "Float", "gtecsfield")
	AssertError(t, errs, "Test.Array", "Test.Array", "Array", "Array", "gtecsfield")

	errs = validate.VarWithValue(1, "", "gtecsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtecsfield")

	// this test is for the WARNING about unforeseen validation issues.
	errs = validate.VarWithValue(test, now, "gtecsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "gtecsfield")
	AssertError(t, errs, "Test.String", "Test.String", "String", "String", "gtecsfield")
	AssertError(t, errs, "Test.Int", "Test.Int", "Int", "Int", "gtecsfield")
	AssertError(t, errs, "Test.Uint", "Test.Uint", "Uint", "Uint", "gtecsfield")
	AssertError(t, errs, "Test.Float", "Test.Float", "Float", "Float", "gtecsfield")
	AssertError(t, errs, "Test.Array", "Test.Array", "Array", "Array", "gtecsfield")

	type Other struct {
		Value string
	}

	type Test2 struct {
		Value Other
		Time  time.Time `validate:"gtecsfield=Value"`
	}

	tst := Test2{
		Value: Other{Value: "StringVal"},
		Time:  then,
	}

	errs = validate.Struct(tst)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test2.Time", "Test2.Time", "Time", "Time", "gtecsfield")

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "gtecsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "gtecsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "gtecsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtecsfield")

	errs = validate.VarWithValue(time.Duration(0), time.Hour, "omitempty,gtecsfield")
	Equal(t, errs, nil)

	// -- Validations for a struct and an inner struct with time.Duration type fields.

	type TimeDurationInner struct {
		Duration time.Duration
	}

	var timeDurationInner *TimeDurationInner

	type TimeDurationTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"gtecsfield=Inner.Duration"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationInner = &TimeDurationInner{time.Hour - time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour + time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "gtecsfield")

	type TimeDurationOmitemptyTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"omitempty,gtecsfield=Inner.Duration"`
	}

	var timeDurationOmitemptyTest *TimeDurationOmitemptyTest
	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationOmitemptyTest = &TimeDurationOmitemptyTest{timeDurationInner, time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestCrossStructGtFieldValidation(t *testing.T) {
	var errs error
	validate := New()
	type Inner struct {
		CreatedAt *time.Time
		String    string
		Int       int
		Uint      uint
		Float     float64
		Array     []string
	}

	type Test struct {
		Inner     *Inner
		CreatedAt *time.Time `validate:"gtcsfield=Inner.CreatedAt"`
		String    string     `validate:"gtcsfield=Inner.String"`
		Int       int        `validate:"gtcsfield=Inner.Int"`
		Uint      uint       `validate:"gtcsfield=Inner.Uint"`
		Float     float64    `validate:"gtcsfield=Inner.Float"`
		Array     []string   `validate:"gtcsfield=Inner.Array"`
	}

	now := time.Now().UTC()
	then := now.Add(time.Hour * -5)
	inner := &Inner{
		CreatedAt: &then,
		String:    "abcd",
		Int:       13,
		Uint:      13,
		Float:     1.13,
		Array:     []string{"val1", "val2"},
	}

	test := &Test{
		Inner:     inner,
		CreatedAt: &now,
		String:    "abcde",
		Int:       14,
		Uint:      14,
		Float:     1.14,
		Array:     []string{"val1", "val2", "val3"},
	}

	errs = validate.Struct(test)
	Equal(t, errs, nil)

	test.CreatedAt = &then
	test.String = "abcd"
	test.Int = 13
	test.Uint = 13
	test.Float = 1.13
	test.Array = []string{"val1", "val2"}

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "gtcsfield")
	AssertError(t, errs, "Test.String", "Test.String", "String", "String", "gtcsfield")
	AssertError(t, errs, "Test.Int", "Test.Int", "Int", "Int", "gtcsfield")
	AssertError(t, errs, "Test.Uint", "Test.Uint", "Uint", "Uint", "gtcsfield")
	AssertError(t, errs, "Test.Float", "Test.Float", "Float", "Float", "gtcsfield")
	AssertError(t, errs, "Test.Array", "Test.Array", "Array", "Array", "gtcsfield")

	errs = validate.VarWithValue(1, "", "gtcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtcsfield")

	// this test is for the WARNING about unforeseen validation issues.
	errs = validate.VarWithValue(test, now, "gtcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "gtcsfield")
	AssertError(t, errs, "Test.String", "Test.String", "String", "String", "gtcsfield")
	AssertError(t, errs, "Test.Int", "Test.Int", "Int", "Int", "gtcsfield")
	AssertError(t, errs, "Test.Uint", "Test.Uint", "Uint", "Uint", "gtcsfield")
	AssertError(t, errs, "Test.Float", "Test.Float", "Float", "Float", "gtcsfield")
	AssertError(t, errs, "Test.Array", "Test.Array", "Array", "Array", "gtcsfield")

	type Other struct {
		Value string
	}

	type Test2 struct {
		Value Other
		Time  time.Time `validate:"gtcsfield=Value"`
	}

	tst := Test2{
		Value: Other{Value: "StringVal"},
		Time:  then,
	}

	errs = validate.Struct(tst)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test2.Time", "Test2.Time", "Time", "Time", "gtcsfield")

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "gtcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "gtcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtcsfield")

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "gtcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtcsfield")

	errs = validate.VarWithValue(time.Duration(0), time.Hour, "omitempty,gtcsfield")
	Equal(t, errs, nil)

	// -- Validations for a struct and an inner struct with time.Duration type fields.

	type TimeDurationInner struct {
		Duration time.Duration
	}

	var timeDurationInner *TimeDurationInner

	type TimeDurationTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"gtcsfield=Inner.Duration"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationInner = &TimeDurationInner{time.Hour - time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "gtcsfield")

	timeDurationInner = &TimeDurationInner{time.Hour + time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "gtcsfield")

	type TimeDurationOmitemptyTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"omitempty,gtcsfield=Inner.Duration"`
	}

	var timeDurationOmitemptyTest *TimeDurationOmitemptyTest
	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationOmitemptyTest = &TimeDurationOmitemptyTest{timeDurationInner, time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestCrossStructNeFieldValidation(t *testing.T) {
	var errs error
	validate := New()
	type Inner struct {
		CreatedAt *time.Time
	}

	type Test struct {
		Inner     *Inner
		CreatedAt *time.Time `validate:"necsfield=Inner.CreatedAt"`
	}

	now := time.Now().UTC()
	then := now.Add(time.Hour * 5)
	inner := &Inner{
		CreatedAt: &then,
	}

	test := &Test{
		Inner:     inner,
		CreatedAt: &now,
	}

	errs = validate.Struct(test)
	Equal(t, errs, nil)

	test.CreatedAt = &then

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "necsfield")

	var j uint64
	var k float64
	var j2 uint64
	var k2 float64
	s := "abcd"
	i := 1
	j = 1
	k = 1.543
	b := true
	arr := []string{"test"}
	s2 := "abcd"
	i2 := 1
	j2 = 1
	k2 = 1.543
	b2 := true
	arr2 := []string{"test"}
	arr3 := []string{"test", "test2"}
	now2 := now

	errs = validate.VarWithValue(s, s2, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	errs = validate.VarWithValue(i2, i, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	errs = validate.VarWithValue(j2, j, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	errs = validate.VarWithValue(k2, k, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	errs = validate.VarWithValue(b2, b, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	errs = validate.VarWithValue(arr2, arr, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	errs = validate.VarWithValue(now2, now, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	errs = validate.VarWithValue(arr3, arr, "necsfield")
	Equal(t, errs, nil)

	type SInner struct {
		Name string
	}

	type TStruct struct {
		Inner     *SInner
		CreatedAt *time.Time `validate:"necsfield=Inner"`
	}

	sinner := &SInner{
		Name: "NAME",
	}

	test2 := &TStruct{
		Inner:     sinner,
		CreatedAt: &now,
	}

	errs = validate.Struct(test2)
	Equal(t, errs, nil)

	test2.Inner = nil
	errs = validate.Struct(test2)
	Equal(t, errs, nil)

	errs = validate.VarWithValue(nil, 1, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "necsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "necsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "necsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "necsfield")

	errs = validate.VarWithValue(time.Duration(0), time.Duration(0), "omitempty,necsfield")
	Equal(t, errs, nil)

	// -- Validations for a struct and an inner struct with time.Duration type fields.

	type TimeDurationInner struct {
		Duration time.Duration
	}

	var timeDurationInner *TimeDurationInner

	type TimeDurationTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"necsfield=Inner.Duration"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationInner = &TimeDurationInner{time.Hour - time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour + time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "necsfield")

	type TimeDurationOmitemptyTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"omitempty,necsfield=Inner.Duration"`
	}

	var timeDurationOmitemptyTest *TimeDurationOmitemptyTest
	timeDurationInner = &TimeDurationInner{time.Duration(0)}
	timeDurationOmitemptyTest = &TimeDurationOmitemptyTest{timeDurationInner, time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestCrossStructEqFieldValidation(t *testing.T) {
	var errs error
	validate := New()
	type Inner struct {
		CreatedAt *time.Time
	}

	type Test struct {
		Inner     *Inner
		CreatedAt *time.Time `validate:"eqcsfield=Inner.CreatedAt"`
	}

	now := time.Now().UTC()
	inner := &Inner{
		CreatedAt: &now,
	}

	test := &Test{
		Inner:     inner,
		CreatedAt: &now,
	}

	errs = validate.Struct(test)
	Equal(t, errs, nil)

	newTime := time.Now().Add(time.Hour).UTC()
	test.CreatedAt = &newTime

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.CreatedAt", "Test.CreatedAt", "CreatedAt", "CreatedAt", "eqcsfield")

	var j uint64
	var k float64
	s := "abcd"
	i := 1
	j = 1
	k = 1.543
	b := true
	arr := []string{"test"}

	var j2 uint64
	var k2 float64
	s2 := "abcd"
	i2 := 1
	j2 = 1
	k2 = 1.543
	b2 := true
	arr2 := []string{"test"}
	arr3 := []string{"test", "test2"}
	now2 := now

	errs = validate.VarWithValue(s, s2, "eqcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(i2, i, "eqcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(j2, j, "eqcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(k2, k, "eqcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(b2, b, "eqcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(arr2, arr, "eqcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(now2, now, "eqcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(arr3, arr, "eqcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqcsfield")

	type SInner struct {
		Name string
	}

	type TStruct struct {
		Inner     *SInner
		CreatedAt *time.Time `validate:"eqcsfield=Inner"`
	}

	sinner := &SInner{
		Name: "NAME",
	}

	test2 := &TStruct{
		Inner:     sinner,
		CreatedAt: &now,
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TStruct.CreatedAt", "TStruct.CreatedAt", "CreatedAt", "CreatedAt", "eqcsfield")

	test2.Inner = nil
	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TStruct.CreatedAt", "TStruct.CreatedAt", "CreatedAt", "CreatedAt", "eqcsfield")

	errs = validate.VarWithValue(nil, 1, "eqcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqcsfield")

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour, "eqcsfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "eqcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqcsfield")

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "eqcsfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqcsfield")

	errs = validate.VarWithValue(time.Duration(0), time.Hour, "omitempty,eqcsfield")
	Equal(t, errs, nil)

	// -- Validations for a struct and an inner struct with time.Duration type fields.

	type TimeDurationInner struct {
		Duration time.Duration
	}

	var timeDurationInner *TimeDurationInner

	type TimeDurationTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"eqcsfield=Inner.Duration"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationInner = &TimeDurationInner{time.Hour - time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "eqcsfield")

	timeDurationInner = &TimeDurationInner{time.Hour + time.Minute}
	timeDurationTest = &TimeDurationTest{timeDurationInner, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "eqcsfield")

	type TimeDurationOmitemptyTest struct {
		Inner    *TimeDurationInner
		Duration time.Duration `validate:"omitempty,eqcsfield=Inner.Duration"`
	}

	var timeDurationOmitemptyTest *TimeDurationOmitemptyTest
	timeDurationInner = &TimeDurationInner{time.Hour}
	timeDurationOmitemptyTest = &TimeDurationOmitemptyTest{timeDurationInner, time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestExistsValidation(t *testing.T) {
	jsonText := "{ \"truthiness2\": true }"
	type Thing struct {
		Truthiness *bool `json:"truthiness" validate:"required"`
	}

	var ting Thing
	err := json.Unmarshal([]byte(jsonText), &ting)
	Equal(t, err, nil)
	NotEqual(t, ting, nil)
	Equal(t, ting.Truthiness, nil)

	validate := New()
	errs := validate.Struct(ting)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Thing.Truthiness", "Thing.Truthiness", "Truthiness", "Truthiness", "required")

	jsonText = "{ \"truthiness\": true }"
	err = json.Unmarshal([]byte(jsonText), &ting)
	Equal(t, err, nil)
	NotEqual(t, ting, nil)
	Equal(t, ting.Truthiness, true)

	errs = validate.Struct(ting)
	Equal(t, errs, nil)
}

func TestSliceMapArrayChanFuncPtrInterfaceRequiredValidation(t *testing.T) {
	var m map[string]string
	validate := New()
	errs := validate.Var(m, "required")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "required")

	m = map[string]string{}
	errs = validate.Var(m, "required")
	Equal(t, errs, nil)

	var arr [5]string
	errs = validate.Var(arr, "required")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "required")

	arr[0] = "ok"
	errs = validate.Var(arr, "required")
	Equal(t, errs, nil)

	var s []string
	errs = validate.Var(s, "required")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "required")

	s = []string{}
	errs = validate.Var(s, "required")
	Equal(t, errs, nil)

	var c chan string
	errs = validate.Var(c, "required")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "required")

	c = make(chan string)
	errs = validate.Var(c, "required")
	Equal(t, errs, nil)

	var tst *int
	errs = validate.Var(tst, "required")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "required")

	one := 1
	tst = &one
	errs = validate.Var(tst, "required")
	Equal(t, errs, nil)

	var iface interface{}

	errs = validate.Var(iface, "required")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "required")

	errs = validate.Var(iface, "omitempty,required")
	Equal(t, errs, nil)

	errs = validate.Var(iface, "")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(nil, iface, "")
	Equal(t, errs, nil)

	var f func(string)
	errs = validate.Var(f, "required")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "required")

	f = func(name string) {}
	errs = validate.Var(f, "required")
	Equal(t, errs, nil)
}

func TestDatePtrValidationIssueValidation(t *testing.T) {
	type Test struct {
		LastViewed *time.Time
		Reminder   *time.Time
	}

	test := &Test{}
	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)
}

func TestCommaAndPipeObfuscationValidation(t *testing.T) {
	s := "My Name Is, |joeybloggs|"
	validate := New()
	errs := validate.Var(s, "excludesall=0x2C")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "excludesall")

	errs = validate.Var(s, "excludesall=0x7C")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "excludesall")
}

func TestBadKeyValidation(t *testing.T) {
	type Test struct {
		Name string `validate:"required, "`
	}

	tst := &Test{
		Name: "test",
	}

	validate := New()

	PanicMatches(t, func() { _ = validate.Struct(tst) }, "Undefined validation function ' ' on field 'Name'")

	type Test2 struct {
		Name string `validate:"required,,len=2"`
	}

	tst2 := &Test2{
		Name: "test",
	}

	PanicMatches(t, func() { _ = validate.Struct(tst2) }, "Invalid validation tag on field 'Name'")
}

func TestArrayDiveValidation(t *testing.T) {
	validate := New()

	arr := []string{"ok", "", "ok"}

	errs := validate.Var(arr, "len=3,dive,required")
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "[1]", "[1]", "[1]", "[1]", "required")

	errs = validate.Var(arr, "len=2,dive,required")
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "", "", "", "", "len")

	type BadDive struct {
		Name string `validate:"dive"`
	}

	bd := &BadDive{
		Name: "TEST",
	}

	PanicMatches(t, func() { _ = validate.Struct(bd) }, "dive error! can't dive on a non slice or map")

	type Test struct {
		Errs []string `validate:"gt=0,dive,required"`
	}

	test := &Test{
		Errs: []string{"ok", "", "ok"},
	}

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "Test.Errs[1]", "Test.Errs[1]", "Errs[1]", "Errs[1]", "required")

	test = &Test{
		Errs: []string{"ok", "ok", ""},
	}

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 1)
	AssertError(t, errs, "Test.Errs[2]", "Test.Errs[2]", "Errs[2]", "Errs[2]", "required")

	type TestMultiDimensional struct {
		Errs [][]string `validate:"gt=0,dive,dive,required"`
	}

	var errArray [][]string
	errArray = append(errArray, []string{"ok", "", ""})
	errArray = append(errArray, []string{"ok", "", ""})

	tm := &TestMultiDimensional{
		Errs: errArray,
	}

	errs = validate.Struct(tm)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 4)
	AssertError(t, errs, "TestMultiDimensional.Errs[0][1]", "TestMultiDimensional.Errs[0][1]", "Errs[0][1]", "Errs[0][1]", "required")
	AssertError(t, errs, "TestMultiDimensional.Errs[0][2]", "TestMultiDimensional.Errs[0][2]", "Errs[0][2]", "Errs[0][2]", "required")
	AssertError(t, errs, "TestMultiDimensional.Errs[1][1]", "TestMultiDimensional.Errs[1][1]", "Errs[1][1]", "Errs[1][1]", "required")
	AssertError(t, errs, "TestMultiDimensional.Errs[1][2]", "TestMultiDimensional.Errs[1][2]", "Errs[1][2]", "Errs[1][2]", "required")

	type Inner struct {
		Name string `validate:"required"`
	}

	type TestMultiDimensionalStructs struct {
		Errs [][]Inner `validate:"gt=0,dive,dive"`
	}

	var errStructArray [][]Inner
	errStructArray = append(errStructArray, []Inner{{"ok"}, {""}, {""}})
	errStructArray = append(errStructArray, []Inner{{"ok"}, {""}, {""}})

	tms := &TestMultiDimensionalStructs{
		Errs: errStructArray,
	}

	errs = validate.Struct(tms)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 4)
	AssertError(t, errs, "TestMultiDimensionalStructs.Errs[0][1].Name", "TestMultiDimensionalStructs.Errs[0][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructs.Errs[0][2].Name", "TestMultiDimensionalStructs.Errs[0][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructs.Errs[1][1].Name", "TestMultiDimensionalStructs.Errs[1][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructs.Errs[1][2].Name", "TestMultiDimensionalStructs.Errs[1][2].Name", "Name", "Name", "required")

	type TestMultiDimensionalStructsPtr struct {
		Errs [][]*Inner `validate:"gt=0,dive,dive"`
	}

	var errStructPtrArray [][]*Inner
	errStructPtrArray = append(errStructPtrArray, []*Inner{{"ok"}, {""}, {""}})
	errStructPtrArray = append(errStructPtrArray, []*Inner{{"ok"}, {""}, {""}})
	errStructPtrArray = append(errStructPtrArray, []*Inner{{"ok"}, {""}, nil})

	tmsp := &TestMultiDimensionalStructsPtr{
		Errs: errStructPtrArray,
	}

	errs = validate.Struct(tmsp)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 5)
	AssertError(t, errs, "TestMultiDimensionalStructsPtr.Errs[0][1].Name", "TestMultiDimensionalStructsPtr.Errs[0][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr.Errs[0][2].Name", "TestMultiDimensionalStructsPtr.Errs[0][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr.Errs[1][1].Name", "TestMultiDimensionalStructsPtr.Errs[1][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr.Errs[1][2].Name", "TestMultiDimensionalStructsPtr.Errs[1][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr.Errs[2][1].Name", "TestMultiDimensionalStructsPtr.Errs[2][1].Name", "Name", "Name", "required")

	// for full test coverage
	s := fmt.Sprint(errs.Error())
	NotEqual(t, s, "")

	type TestMultiDimensionalStructsPtr2 struct {
		Errs [][]*Inner `validate:"gt=0,dive,dive,required"`
	}

	var errStructPtr2Array [][]*Inner
	errStructPtr2Array = append(errStructPtr2Array, []*Inner{{"ok"}, {""}, {""}})
	errStructPtr2Array = append(errStructPtr2Array, []*Inner{{"ok"}, {""}, {""}})
	errStructPtr2Array = append(errStructPtr2Array, []*Inner{{"ok"}, {""}, nil})

	tmsp2 := &TestMultiDimensionalStructsPtr2{
		Errs: errStructPtr2Array,
	}

	errs = validate.Struct(tmsp2)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 6)
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[0][1].Name", "TestMultiDimensionalStructsPtr2.Errs[0][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[0][2].Name", "TestMultiDimensionalStructsPtr2.Errs[0][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[1][1].Name", "TestMultiDimensionalStructsPtr2.Errs[1][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[1][2].Name", "TestMultiDimensionalStructsPtr2.Errs[1][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[2][1].Name", "TestMultiDimensionalStructsPtr2.Errs[2][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr2.Errs[2][2]", "TestMultiDimensionalStructsPtr2.Errs[2][2]", "Errs[2][2]", "Errs[2][2]", "required")

	type TestMultiDimensionalStructsPtr3 struct {
		Errs [][]*Inner `validate:"gt=0,dive,dive,omitempty"`
	}

	var errStructPtr3Array [][]*Inner
	errStructPtr3Array = append(errStructPtr3Array, []*Inner{{"ok"}, {""}, {""}})
	errStructPtr3Array = append(errStructPtr3Array, []*Inner{{"ok"}, {""}, {""}})
	errStructPtr3Array = append(errStructPtr3Array, []*Inner{{"ok"}, {""}, nil})

	tmsp3 := &TestMultiDimensionalStructsPtr3{
		Errs: errStructPtr3Array,
	}

	errs = validate.Struct(tmsp3)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 5)
	AssertError(t, errs, "TestMultiDimensionalStructsPtr3.Errs[0][1].Name", "TestMultiDimensionalStructsPtr3.Errs[0][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr3.Errs[0][2].Name", "TestMultiDimensionalStructsPtr3.Errs[0][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr3.Errs[1][1].Name", "TestMultiDimensionalStructsPtr3.Errs[1][1].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr3.Errs[1][2].Name", "TestMultiDimensionalStructsPtr3.Errs[1][2].Name", "Name", "Name", "required")
	AssertError(t, errs, "TestMultiDimensionalStructsPtr3.Errs[2][1].Name", "TestMultiDimensionalStructsPtr3.Errs[2][1].Name", "Name", "Name", "required")

	type TestMultiDimensionalTimeTime struct {
		Errs [][]*time.Time `validate:"gt=0,dive,dive,required"`
	}

	var errTimePtr3Array [][]*time.Time
	t1 := time.Now().UTC()
	t2 := time.Now().UTC()
	t3 := time.Now().UTC().Add(time.Hour * 24)
	errTimePtr3Array = append(errTimePtr3Array, []*time.Time{&t1, &t2, &t3})
	errTimePtr3Array = append(errTimePtr3Array, []*time.Time{&t1, &t2, nil})
	errTimePtr3Array = append(errTimePtr3Array, []*time.Time{&t1, nil, nil})

	tmtp3 := &TestMultiDimensionalTimeTime{
		Errs: errTimePtr3Array,
	}

	errs = validate.Struct(tmtp3)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 3)
	AssertError(t, errs, "TestMultiDimensionalTimeTime.Errs[1][2]", "TestMultiDimensionalTimeTime.Errs[1][2]", "Errs[1][2]", "Errs[1][2]", "required")
	AssertError(t, errs, "TestMultiDimensionalTimeTime.Errs[2][1]", "TestMultiDimensionalTimeTime.Errs[2][1]", "Errs[2][1]", "Errs[2][1]", "required")
	AssertError(t, errs, "TestMultiDimensionalTimeTime.Errs[2][2]", "TestMultiDimensionalTimeTime.Errs[2][2]", "Errs[2][2]", "Errs[2][2]", "required")

	type TestMultiDimensionalTimeTime2 struct {
		Errs [][]*time.Time `validate:"gt=0,dive,dive,required"`
	}

	var errTimeArray [][]*time.Time
	t1 = time.Now().UTC()
	t2 = time.Now().UTC()
	t3 = time.Now().UTC().Add(time.Hour * 24)
	errTimeArray = append(errTimeArray, []*time.Time{&t1, &t2, &t3})
	errTimeArray = append(errTimeArray, []*time.Time{&t1, &t2, nil})
	errTimeArray = append(errTimeArray, []*time.Time{&t1, nil, nil})

	tmtp := &TestMultiDimensionalTimeTime2{
		Errs: errTimeArray,
	}

	errs = validate.Struct(tmtp)
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 3)
	AssertError(t, errs, "TestMultiDimensionalTimeTime2.Errs[1][2]", "TestMultiDimensionalTimeTime2.Errs[1][2]", "Errs[1][2]", "Errs[1][2]", "required")
	AssertError(t, errs, "TestMultiDimensionalTimeTime2.Errs[2][1]", "TestMultiDimensionalTimeTime2.Errs[2][1]", "Errs[2][1]", "Errs[2][1]", "required")
	AssertError(t, errs, "TestMultiDimensionalTimeTime2.Errs[2][2]", "TestMultiDimensionalTimeTime2.Errs[2][2]", "Errs[2][2]", "Errs[2][2]", "required")
}

func TestNilStructPointerValidation(t *testing.T) {
	type Inner struct {
		Data string
	}

	type Outer struct {
		Inner *Inner `validate:"omitempty"`
	}

	inner := &Inner{
		Data: "test",
	}

	outer := &Outer{
		Inner: inner,
	}

	validate := New()
	errs := validate.Struct(outer)
	Equal(t, errs, nil)

	outer = &Outer{
		Inner: nil,
	}

	errs = validate.Struct(outer)
	Equal(t, errs, nil)

	type Inner2 struct {
		Data string
	}

	type Outer2 struct {
		Inner2 *Inner2 `validate:"required"`
	}

	inner2 := &Inner2{
		Data: "test",
	}

	outer2 := &Outer2{
		Inner2: inner2,
	}

	errs = validate.Struct(outer2)
	Equal(t, errs, nil)

	outer2 = &Outer2{
		Inner2: nil,
	}

	errs = validate.Struct(outer2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Outer2.Inner2", "Outer2.Inner2", "Inner2", "Inner2", "required")

	type Inner3 struct {
		Data string
	}

	type Outer3 struct {
		Inner3 *Inner3
	}

	inner3 := &Inner3{
		Data: "test",
	}

	outer3 := &Outer3{
		Inner3: inner3,
	}

	errs = validate.Struct(outer3)
	Equal(t, errs, nil)

	type Inner4 struct {
		Data string
	}

	type Outer4 struct {
		Inner4 *Inner4 `validate:"-"`
	}

	inner4 := &Inner4{
		Data: "test",
	}

	outer4 := &Outer4{
		Inner4: inner4,
	}

	errs = validate.Struct(outer4)
	Equal(t, errs, nil)
}

func TestExcludesRuneValidation(t *testing.T) {
	tests := []struct {
		Value       string `validate:"excludesrune=☻"`
		Tag         string
		ExpectedNil bool
	}{
		{Value: "a☺b☻c☹d", Tag: "excludesrune=☻", ExpectedNil: false},
		{Value: "abcd", Tag: "excludesrune=☻", ExpectedNil: true},
	}

	validate := New()
	for i, s := range tests {
		errs := validate.Var(s.Value, s.Tag)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}

		errs = validate.Struct(s)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}
	}
}

func TestExcludesAllValidation(t *testing.T) {
	tests := []struct {
		Value       string `validate:"excludesall=@!{}[]"`
		Tag         string
		ExpectedNil bool
	}{
		{Value: "abcd@!jfk", Tag: "excludesall=@!{}[]", ExpectedNil: false},
		{Value: "abcdefg", Tag: "excludesall=@!{}[]", ExpectedNil: true},
	}

	validate := New()
	for i, s := range tests {
		errs := validate.Var(s.Value, s.Tag)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}

		errs = validate.Struct(s)
		if (s.ExpectedNil && errs != nil) || (!s.ExpectedNil && errs == nil) {
			t.Fatalf("Index: %d failed Error: %s", i, errs)
		}
	}

	username := "joeybloggs "
	errs := validate.Var(username, "excludesall=@ ")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "excludesall")

	excluded := ","
	errs = validate.Var(excluded, "excludesall=!@#$%^&*()_+.0x2C?")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "excludesall")

	excluded = "="
	errs = validate.Var(excluded, "excludesall=!@#$%^&*()_+.0x2C=?")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "excludesall")
}

func TestIsNeFieldValidation(t *testing.T) {
	var errs error
	var j uint64
	var k float64
	validate := New()
	s := "abcd"
	i := 1
	j = 1
	k = 1.543
	b := true
	arr := []string{"test"}
	now := time.Now().UTC()

	var j2 uint64
	var k2 float64
	s2 := "abcdef"
	i2 := 3
	j2 = 2
	k2 = 1.5434456
	b2 := false
	arr2 := []string{"test", "test2"}
	arr3 := []string{"test"}
	now2 := now

	errs = validate.VarWithValue(s, s2, "nefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(i2, i, "nefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(j2, j, "nefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(k2, k, "nefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(b2, b, "nefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(arr2, arr, "nefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(now2, now, "nefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "nefield")

	errs = validate.VarWithValue(arr3, arr, "nefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "nefield")

	type Test struct {
		Start *time.Time `validate:"nefield=End"`
		End   *time.Time
	}

	sv := &Test{
		Start: &now,
		End:   &now,
	}

	errs = validate.Struct(sv)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.Start", "Test.Start", "Start", "Start", "nefield")

	now3 := time.Now().Add(time.Hour).UTC()

	sv = &Test{
		Start: &now,
		End:   &now3,
	}

	errs = validate.Struct(sv)
	Equal(t, errs, nil)

	errs = validate.VarWithValue(nil, 1, "nefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "nefield")

	errs = validate.VarWithValue(sv, now, "nefield")
	Equal(t, errs, nil)

	type Test2 struct {
		Start *time.Time `validate:"nefield=NonExistantField"`
		End   *time.Time
	}

	sv2 := &Test2{
		Start: &now,
		End:   &now,
	}

	errs = validate.Struct(sv2)
	Equal(t, errs, nil)

	type Other struct {
		Value string
	}

	type Test3 struct {
		Value Other
		Time  time.Time `validate:"nefield=Value"`
	}

	tst := Test3{
		Value: Other{Value: "StringVal"},
		Time:  now,
	}

	errs = validate.Struct(tst)
	Equal(t, errs, nil)

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "nefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "nefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "nefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "nefield")

	errs = validate.VarWithValue(time.Duration(0), time.Duration(0), "omitempty,nefield")
	Equal(t, errs, nil)

	// -- Validations for a struct with time.Duration type fields.

	type TimeDurationTest struct {
		First  time.Duration `validate:"nefield=Second"`
		Second time.Duration
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "nefield")

	type TimeDurationOmitemptyTest struct {
		First  time.Duration `validate:"omitempty,nefield=Second"`
		Second time.Duration
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0), time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestIsNeValidation(t *testing.T) {
	var errs error
	var j uint64
	var k float64
	validate := New()
	s := "abcdef"
	i := 3
	j = 2
	k = 1.5434
	arr := []string{"test"}
	now := time.Now().UTC()

	errs = validate.Var(s, "ne=abcd")
	Equal(t, errs, nil)

	errs = validate.Var(i, "ne=1")
	Equal(t, errs, nil)

	errs = validate.Var(j, "ne=1")
	Equal(t, errs, nil)

	errs = validate.Var(k, "ne=1.543")
	Equal(t, errs, nil)

	errs = validate.Var(arr, "ne=2")
	Equal(t, errs, nil)

	errs = validate.Var(arr, "ne=1")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ne")

	PanicMatches(t, func() { _ = validate.Var(now, "ne=now") }, "Bad field type time.Time")

	// Tests for time.Duration type.

	// -- Validations for a variable of time.Duration type.

	errs = validate.Var(time.Hour-time.Minute, "ne=1h")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour+time.Minute, "ne=1h")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour, "ne=1h")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ne")

	errs = validate.Var(time.Duration(0), "omitempty,ne=0")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.

	type TimeDurationTest struct {
		Duration time.Duration `validate:"ne=1h"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "ne")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,ne=0"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestIsNeIgnoreCaseValidation(t *testing.T) {
	var errs error
	validate := New()
	s := "abcd"
	now := time.Now()
	errs = validate.Var(s, "ne_ignore_case=efgh")
	Equal(t, errs, nil)

	errs = validate.Var(s, "ne_ignore_case=AbCd")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ne_ignore_case")

	PanicMatches(
		t, func() { _ = validate.Var(now, "ne_ignore_case=abcd") }, "Bad field type time.Time",
	)
}

func TestIsEqFieldValidation(t *testing.T) {
	var errs error
	var j uint64
	var k float64
	validate := New()
	s := "abcd"
	i := 1
	j = 1
	k = 1.543
	b := true
	arr := []string{"test"}
	now := time.Now().UTC()

	var j2 uint64
	var k2 float64
	s2 := "abcd"
	i2 := 1
	j2 = 1
	k2 = 1.543
	b2 := true
	arr2 := []string{"test"}
	arr3 := []string{"test", "test2"}
	now2 := now

	errs = validate.VarWithValue(s, s2, "eqfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(i2, i, "eqfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(j2, j, "eqfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(k2, k, "eqfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(b2, b, "eqfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(arr2, arr, "eqfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(now2, now, "eqfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(arr3, arr, "eqfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqfield")

	type Test struct {
		Start *time.Time `validate:"eqfield=End"`
		End   *time.Time
	}

	sv := &Test{
		Start: &now,
		End:   &now,
	}

	errs = validate.Struct(sv)
	Equal(t, errs, nil)

	now3 := time.Now().Add(time.Hour).UTC()
	sv = &Test{
		Start: &now,
		End:   &now3,
	}

	errs = validate.Struct(sv)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.Start", "Test.Start", "Start", "Start", "eqfield")

	errs = validate.VarWithValue(nil, 1, "eqfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqfield")

	channel := make(chan string)
	errs = validate.VarWithValue(5, channel, "eqfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqfield")

	errs = validate.VarWithValue(5, now, "eqfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqfield")

	type Test2 struct {
		Start *time.Time `validate:"eqfield=NonExistantField"`
		End   *time.Time
	}

	sv2 := &Test2{
		Start: &now,
		End:   &now,
	}

	errs = validate.Struct(sv2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test2.Start", "Test2.Start", "Start", "Start", "eqfield")

	type Inner struct {
		Name string
	}

	type TStruct struct {
		Inner     *Inner
		CreatedAt *time.Time `validate:"eqfield=Inner"`
	}

	inner := &Inner{
		Name: "NAME",
	}

	test := &TStruct{
		Inner:     inner,
		CreatedAt: &now,
	}

	errs = validate.Struct(test)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TStruct.CreatedAt", "TStruct.CreatedAt", "CreatedAt", "CreatedAt", "eqfield")

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour, "eqfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "eqfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqfield")

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "eqfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eqfield")

	errs = validate.VarWithValue(time.Duration(0), time.Hour, "omitempty,eqfield")
	Equal(t, errs, nil)

	// -- Validations for a struct with time.Duration type fields.

	type TimeDurationTest struct {
		First  time.Duration `validate:"eqfield=Second"`
		Second time.Duration
	}
	var timeDurationTest *TimeDurationTest

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "eqfield")

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "eqfield")

	type TimeDurationOmitemptyTest struct {
		First  time.Duration `validate:"omitempty,eqfield=Second"`
		Second time.Duration
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0), time.Hour}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestIsEqFieldValidationWithAliasTime(t *testing.T) {
	var errs error
	validate := New()

	type CustomTime time.Time

	type Test struct {
		Start CustomTime `validate:"eqfield=End"`
		End   *time.Time
	}

	now := time.Now().UTC()
	sv := &Test{
		Start: CustomTime(now),
		End:   &now,
	}

	errs = validate.Struct(sv)
	Equal(t, errs, nil)
}

func TestIsEqValidation(t *testing.T) {
	var errs error
	var j uint64
	var k float64
	validate := New()
	s := "abcd"
	i := 1
	j = 1
	k = 1.543
	arr := []string{"test"}
	now := time.Now().UTC()

	errs = validate.Var(s, "eq=abcd")
	Equal(t, errs, nil)

	errs = validate.Var(i, "eq=1")
	Equal(t, errs, nil)

	errs = validate.Var(j, "eq=1")
	Equal(t, errs, nil)

	errs = validate.Var(k, "eq=1.543")
	Equal(t, errs, nil)

	errs = validate.Var(arr, "eq=1")
	Equal(t, errs, nil)

	errs = validate.Var(arr, "eq=2")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eq")

	PanicMatches(t, func() { _ = validate.Var(now, "eq=now") }, "Bad field type time.Time")

	// Tests for time.Duration type.

	// -- Validations for a variable of time.Duration type.

	errs = validate.Var(time.Hour, "eq=1h")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-time.Minute, "eq=1h")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eq")

	errs = validate.Var(time.Hour+time.Minute, "eq=1h")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "eq")

	errs = validate.Var(time.Duration(0), "omitempty,eq=1h")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.

	type TimeDurationTest struct {
		Duration time.Duration `validate:"eq=1h"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "eq")

	timeDurationTest = &TimeDurationTest{time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "eq")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,eq=1h"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestIsEqIgnoreCaseValidation(t *testing.T) {
	var errs error
	validate := New()
	s := "abcd"
	now := time.Now()
	errs = validate.Var(s, "eq_ignore_case=abcd")
	Equal(t, errs, nil)

	errs = validate.Var(s, "eq_ignore_case=AbCd")
	Equal(t, errs, nil)

	PanicMatches(
		t, func() { _ = validate.Var(now, "eq_ignore_case=abcd") }, "Bad field type time.Time",
	)
}

func TestOneOfValidation(t *testing.T) {
	validate := New()
	passSpecs := []struct {
		f interface{}
		t string
	}{
		{f: "red", t: "oneof=red green"},
		{f: "green", t: "oneof=red green"},
		{f: "red green", t: "oneof='red green' blue"},
		{f: "blue", t: "oneof='red green' blue"},
		{f: 5, t: "oneof=5 6"},
		{f: 6, t: "oneof=5 6"},
		{f: int8(6), t: "oneof=5 6"},
		{f: int16(6), t: "oneof=5 6"},
		{f: int32(6), t: "oneof=5 6"},
		{f: int64(6), t: "oneof=5 6"},
		{f: uint(6), t: "oneof=5 6"},
		{f: uint8(6), t: "oneof=5 6"},
		{f: uint16(6), t: "oneof=5 6"},
		{f: uint32(6), t: "oneof=5 6"},
		{f: uint64(6), t: "oneof=5 6"},
	}

	for _, spec := range passSpecs {
		t.Logf("%#v", spec)
		errs := validate.Var(spec.f, spec.t)
		Equal(t, errs, nil)
	}

	failSpecs := []struct {
		f interface{}
		t string
	}{
		{f: "", t: "oneof=red green"},
		{f: "yellow", t: "oneof=red green"},
		{f: "green", t: "oneof='red green' blue"},
		{f: 5, t: "oneof=red green"},
		{f: 6, t: "oneof=red green"},
		{f: 6, t: "oneof=7"},
		{f: uint(6), t: "oneof=7"},
		{f: int8(5), t: "oneof=red green"},
		{f: int16(5), t: "oneof=red green"},
		{f: int32(5), t: "oneof=red green"},
		{f: int64(5), t: "oneof=red green"},
		{f: uint(5), t: "oneof=red green"},
		{f: uint8(5), t: "oneof=red green"},
		{f: uint16(5), t: "oneof=red green"},
		{f: uint32(5), t: "oneof=red green"},
		{f: uint64(5), t: "oneof=red green"},
	}

	for _, spec := range failSpecs {
		t.Logf("%#v", spec)
		errs := validate.Var(spec.f, spec.t)
		AssertError(t, errs, "", "", "", "", "oneof")
	}

	PanicMatches(t, func() {
		_ = validate.Var(3.14, "oneof=red green")
	}, "Bad field type float64")
}

func TestOneOfCIValidation(t *testing.T) {
	validate := New()
	passSpecs := []struct {
		f interface{}
		t string
	}{
		{f: "red", t: "oneofci=RED GREEN"},
		{f: "RED", t: "oneofci=red green"},
		{f: "red", t: "oneofci=red green"},
		{f: "RED", t: "oneofci=RED GREEN"},
		{f: "green", t: "oneofci=red green"},
		{f: "red green", t: "oneofci='red green' blue"},
		{f: "blue", t: "oneofci='red green' blue"},
		{f: "GREEN", t: "oneofci=Red Green"},
		{f: "ReD", t: "oneofci=RED GREEN"},
		{f: "gReEn", t: "oneofci=rEd GrEeN"},
		{f: "RED GREEN", t: "oneofci='red green' blue"},
		{f: "red Green", t: "oneofci='RED GREEN' Blue"},
		{f: "Red green", t: "oneofci='Red Green' BLUE"},
		{f: "rEd GrEeN", t: "oneofci='ReD gReEn' BlUe"},
		{f: "BLUE", t: "oneofci='Red Green' BLUE"},
		{f: "BlUe", t: "oneofci='RED GREEN' Blue"},
		{f: "bLuE", t: "oneofci='red green' BLUE"},
	}

	for _, spec := range passSpecs {
		t.Logf("%#v", spec)
		errs := validate.Var(spec.f, spec.t)
		Equal(t, errs, nil)
	}

	failSpecs := []struct {
		f interface{}
		t string
	}{
		{f: "", t: "oneofci=red green"},
		{f: "yellow", t: "oneofci=red green"},
		{f: "green", t: "oneofci='red green' blue"},
	}

	for _, spec := range failSpecs {
		t.Logf("%#v", spec)
		errs := validate.Var(spec.f, spec.t)
		AssertError(t, errs, "", "", "", "", "oneofci")
	}

	panicSpecs := []struct {
		f interface{}
		t string
	}{
		{f: 3.14, t: "oneofci=red green"},
		{f: 5, t: "oneofci=red green"},
		{f: uint(6), t: "oneofci=7"},
		{f: int8(5), t: "oneofci=red green"},
		{f: int16(5), t: "oneofci=red green"},
		{f: int32(5), t: "oneofci=red green"},
		{f: int64(5), t: "oneofci=red green"},
		{f: uint(5), t: "oneofci=red green"},
		{f: uint8(5), t: "oneofci=red green"},
		{f: uint16(5), t: "oneofci=red green"},
		{f: uint32(5), t: "oneofci=red green"},
		{f: uint64(5), t: "oneofci=red green"},
	}

	var panicCount int
	for _, spec := range panicSpecs {
		t.Logf("%#v", spec)
		PanicMatches(t, func() {
			_ = validate.Var(spec.f, spec.t)
		}, fmt.Sprintf("Bad field type %T", spec.f))
		panicCount++
	}

	Equal(t, panicCount, len(panicSpecs))
}

func TestBase32Validation(t *testing.T) {
	validate := New()
	s := "ABCD2345"
	errs := validate.Var(s, "base32")
	Equal(t, errs, nil)

	s = "AB======"
	errs = validate.Var(s, "base32")
	Equal(t, errs, nil)

	s = "ABCD2==="
	errs = validate.Var(s, "base32")
	Equal(t, errs, nil)

	s = "ABCD===="
	errs = validate.Var(s, "base32")
	Equal(t, errs, nil)

	s = "ABCD234="
	errs = validate.Var(s, "base32")
	Equal(t, errs, nil)

	s = ""
	errs = validate.Var(s, "base32")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "base32")

	s = "ABCabc1890== foo bar"
	errs = validate.Var(s, "base32")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "base32")
}

func TestBase64Validation(t *testing.T) {
	validate := New()
	s := "dW5pY29ybg=="
	errs := validate.Var(s, "base64")
	Equal(t, errs, nil)

	s = "dGhpIGlzIGEgdGVzdCBiYXNlNjQ="
	errs = validate.Var(s, "base64")
	Equal(t, errs, nil)

	s = ""
	errs = validate.Var(s, "base64")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "base64")

	s = "dW5pY29ybg== foo bar"
	errs = validate.Var(s, "base64")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "base64")
}

func TestNoStructLevelValidation(t *testing.T) {
	type Inner struct {
		Test string `validate:"len=5"`
	}

	type Outer struct {
		InnerStruct    Inner  `validate:"required,nostructlevel"`
		InnerStructPtr *Inner `validate:"required,nostructlevel"`
	}

	outer := &Outer{
		InnerStructPtr: nil,
		InnerStruct:    Inner{},
	}

	// test with struct required failing on
	validate := New(WithRequiredStructEnabled())
	errs := validate.Struct(outer)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Outer.InnerStruct", "Outer.InnerStruct", "InnerStruct", "InnerStruct", "required")
	AssertError(t, errs, "Outer.InnerStructPtr", "Outer.InnerStructPtr", "InnerStructPtr", "InnerStructPtr", "required")

	inner := Inner{
		Test: "1234",
	}

	outer = &Outer{
		InnerStruct:    inner,
		InnerStructPtr: &inner,
	}

	errs = validate.Struct(outer)
	Equal(t, errs, nil)

	// test with struct required failing off

	outer = &Outer{
		InnerStructPtr: nil,
		InnerStruct:    Inner{},
	}
	validate = New()

	errs = validate.Struct(outer)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Outer.InnerStructPtr", "Outer.InnerStructPtr", "InnerStructPtr", "InnerStructPtr", "required")

	inner = Inner{
		Test: "1234",
	}

	outer = &Outer{
		InnerStruct:    inner,
		InnerStructPtr: &inner,
	}

	errs = validate.Struct(outer)
	Equal(t, errs, nil)
}

func TestStructOnlyValidation(t *testing.T) {
	type Inner struct {
		Test string `validate:"len=5"`
	}

	type Outer struct {
		InnerStruct    Inner  `validate:"required,structonly"`
		InnerStructPtr *Inner `validate:"required,structonly"`
	}

	outer := &Outer{
		InnerStruct:    Inner{},
		InnerStructPtr: nil,
	}

	// without required struct on
	validate := New()
	errs := validate.Struct(outer)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Outer.InnerStructPtr", "Outer.InnerStructPtr", "InnerStructPtr", "InnerStructPtr", "required")

	// with required struct on
	validate.requiredStructEnabled = true

	errs = validate.Struct(outer)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Outer.InnerStruct", "Outer.InnerStruct", "InnerStruct", "InnerStruct", "required")
	AssertError(t, errs, "Outer.InnerStructPtr", "Outer.InnerStructPtr", "InnerStructPtr", "InnerStructPtr", "required")

	inner := Inner{
		Test: "1234",
	}

	outer = &Outer{
		InnerStruct:    inner,
		InnerStructPtr: &inner,
	}

	errs = validate.Struct(outer)
	Equal(t, errs, nil)

	// Address houses a users address information
	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
		Planet string `validate:"required"`
		Phone  string `validate:"required"`
	}

	type User struct {
		FirstName      string     `json:"fname"`
		LastName       string     `json:"lname"`
		Age            uint8      `validate:"gte=0,lte=130"`
		Number         string     `validate:"required,e164"`
		Email          string     `validate:"required,email"`
		FavouriteColor string     `validate:"hexcolor|rgb|rgba"`
		Addresses      []*Address `validate:"required"`   // a person can have a home and cottage...
		Address        Address    `validate:"structonly"` // a person can have a home and cottage...
	}

	address := &Address{
		Street: "Eavesdown Docks",
		Planet: "Persphone",
		Phone:  "none",
		City:   "Unknown",
	}

	user := &User{
		FirstName:      "",
		LastName:       "",
		Age:            45,
		Number:         "+1123456789",
		Email:          "Badger.Smith@gmail.com",
		FavouriteColor: "#000",
		Addresses:      []*Address{address},
		Address: Address{
			// Street: "Eavesdown Docks",
			Planet: "Persphone",
			Phone:  "none",
			City:   "Unknown",
		},
	}

	errs = validate.Struct(user)
	Equal(t, errs, nil)
}

func TestGtField(t *testing.T) {
	var errs error
	validate := New()
	type TimeTest struct {
		Start *time.Time `validate:"required,gt"`
		End   *time.Time `validate:"required,gt,gtfield=Start"`
	}

	now := time.Now()
	start := now.Add(time.Hour * 24)
	end := start.Add(time.Hour * 24)

	timeTest := &TimeTest{
		Start: &start,
		End:   &end,
	}

	errs = validate.Struct(timeTest)
	Equal(t, errs, nil)

	timeTest = &TimeTest{
		Start: &end,
		End:   &start,
	}

	errs = validate.Struct(timeTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeTest.End", "TimeTest.End", "End", "End", "gtfield")

	errs = validate.VarWithValue(&end, &start, "gtfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(&start, &end, "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	errs = validate.VarWithValue(&end, &start, "gtfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(&timeTest, &end, "gtfield")
	NotEqual(t, errs, nil)

	errs = validate.VarWithValue("test bigger", "test", "gtfield")
	Equal(t, errs, nil)

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "gtfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	errs = validate.VarWithValue(time.Duration(0), time.Hour, "omitempty,gtfield")
	Equal(t, errs, nil)

	// -- Validations for a struct with time.Duration type fields.

	type TimeDurationTest struct {
		First  time.Duration `validate:"gtfield=Second"`
		Second time.Duration
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "gtfield")

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "gtfield")

	type TimeDurationOmitemptyTest struct {
		First  time.Duration `validate:"omitempty,gtfield=Second"`
		Second time.Duration
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0), time.Hour}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)

	// Tests for Ints types.

	type IntTest struct {
		Val1 int `validate:"required"`
		Val2 int `validate:"required,gtfield=Val1"`
	}

	intTest := &IntTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(intTest)
	Equal(t, errs, nil)

	intTest = &IntTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(intTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "IntTest.Val2", "IntTest.Val2", "Val2", "Val2", "gtfield")

	errs = validate.VarWithValue(int(5), int(1), "gtfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(int(1), int(5), "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	type UIntTest struct {
		Val1 uint `validate:"required"`
		Val2 uint `validate:"required,gtfield=Val1"`
	}

	uIntTest := &UIntTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(uIntTest)
	Equal(t, errs, nil)

	uIntTest = &UIntTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(uIntTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "UIntTest.Val2", "UIntTest.Val2", "Val2", "Val2", "gtfield")

	errs = validate.VarWithValue(uint(5), uint(1), "gtfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(uint(1), uint(5), "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	type FloatTest struct {
		Val1 float64 `validate:"required"`
		Val2 float64 `validate:"required,gtfield=Val1"`
	}

	floatTest := &FloatTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(floatTest)
	Equal(t, errs, nil)

	floatTest = &FloatTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(floatTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "FloatTest.Val2", "FloatTest.Val2", "Val2", "Val2", "gtfield")

	errs = validate.VarWithValue(float32(5), float32(1), "gtfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(float32(1), float32(5), "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	errs = validate.VarWithValue(nil, 1, "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	errs = validate.VarWithValue(5, "T", "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	errs = validate.VarWithValue(5, start, "gtfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtfield")

	type TimeTest2 struct {
		Start *time.Time `validate:"required"`
		End   *time.Time `validate:"required,gtfield=NonExistantField"`
	}

	timeTest2 := &TimeTest2{
		Start: &start,
		End:   &end,
	}

	errs = validate.Struct(timeTest2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeTest2.End", "TimeTest2.End", "End", "End", "gtfield")

	type Other struct {
		Value string
	}

	type Test struct {
		Value Other
		Time  time.Time `validate:"gtfield=Value"`
	}

	tst := Test{
		Value: Other{Value: "StringVal"},
		Time:  end,
	}

	errs = validate.Struct(tst)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.Time", "Test.Time", "Time", "Time", "gtfield")
}

func TestLtField(t *testing.T) {
	var errs error
	validate := New()
	type TimeTest struct {
		Start *time.Time `validate:"required,lt,ltfield=End"`
		End   *time.Time `validate:"required,lt"`
	}

	now := time.Now()
	start := now.Add(time.Hour * 24 * -1 * 2)
	end := start.Add(time.Hour * 24)

	timeTest := &TimeTest{
		Start: &start,
		End:   &end,
	}

	errs = validate.Struct(timeTest)
	Equal(t, errs, nil)

	timeTest = &TimeTest{
		Start: &end,
		End:   &start,
	}

	errs = validate.Struct(timeTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeTest.Start", "TimeTest.Start", "Start", "Start", "ltfield")

	errs = validate.VarWithValue(&start, &end, "ltfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(&end, &start, "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	errs = validate.VarWithValue(&end, timeTest, "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	errs = validate.VarWithValue("tes", "test", "ltfield")
	Equal(t, errs, nil)

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "ltfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	errs = validate.VarWithValue(time.Duration(0), -time.Minute, "omitempty,ltfield")
	Equal(t, errs, nil)

	// -- Validations for a struct with time.Duration type fields.

	type TimeDurationTest struct {
		First  time.Duration `validate:"ltfield=Second"`
		Second time.Duration
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "ltfield")

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "ltfield")

	type TimeDurationOmitemptyTest struct {
		First  time.Duration `validate:"omitempty,ltfield=Second"`
		Second time.Duration
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0), -time.Minute}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)

	// Tests for Ints types.

	type IntTest struct {
		Val1 int `validate:"required"`
		Val2 int `validate:"required,ltfield=Val1"`
	}

	intTest := &IntTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(intTest)
	Equal(t, errs, nil)

	intTest = &IntTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(intTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "IntTest.Val2", "IntTest.Val2", "Val2", "Val2", "ltfield")

	errs = validate.VarWithValue(int(1), int(5), "ltfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(int(5), int(1), "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	type UIntTest struct {
		Val1 uint `validate:"required"`
		Val2 uint `validate:"required,ltfield=Val1"`
	}

	uIntTest := &UIntTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(uIntTest)
	Equal(t, errs, nil)

	uIntTest = &UIntTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(uIntTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "UIntTest.Val2", "UIntTest.Val2", "Val2", "Val2", "ltfield")

	errs = validate.VarWithValue(uint(1), uint(5), "ltfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(uint(5), uint(1), "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	type FloatTest struct {
		Val1 float64 `validate:"required"`
		Val2 float64 `validate:"required,ltfield=Val1"`
	}

	floatTest := &FloatTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(floatTest)
	Equal(t, errs, nil)

	floatTest = &FloatTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(floatTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "FloatTest.Val2", "FloatTest.Val2", "Val2", "Val2", "ltfield")

	errs = validate.VarWithValue(float32(1), float32(5), "ltfield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(float32(5), float32(1), "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	errs = validate.VarWithValue(nil, 5, "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	errs = validate.VarWithValue(1, "T", "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	errs = validate.VarWithValue(1, end, "ltfield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltfield")

	type TimeTest2 struct {
		Start *time.Time `validate:"required"`
		End   *time.Time `validate:"required,ltfield=NonExistantField"`
	}

	timeTest2 := &TimeTest2{
		Start: &end,
		End:   &start,
	}

	errs = validate.Struct(timeTest2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeTest2.End", "TimeTest2.End", "End", "End", "ltfield")
}

func TestFieldContains(t *testing.T) {
	validate := New()
	type StringTest struct {
		Foo string `validate:"fieldcontains=Bar"`
		Bar string
	}

	stringTest := &StringTest{
		Foo: "foobar",
		Bar: "bar",
	}

	errs := validate.Struct(stringTest)
	Equal(t, errs, nil)

	stringTest = &StringTest{
		Foo: "foo",
		Bar: "bar",
	}

	errs = validate.Struct(stringTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "StringTest.Foo", "StringTest.Foo", "Foo", "Foo", "fieldcontains")

	errs = validate.VarWithValue("foo", "bar", "fieldcontains")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "fieldcontains")

	errs = validate.VarWithValue("bar", "foobarfoo", "fieldcontains")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "fieldcontains")

	errs = validate.VarWithValue("foobarfoo", "bar", "fieldcontains")
	Equal(t, errs, nil)

	type StringTestMissingField struct {
		Foo string `validate:"fieldcontains=Bar"`
	}

	stringTestMissingField := &StringTestMissingField{
		Foo: "foo",
	}

	errs = validate.Struct(stringTestMissingField)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "StringTestMissingField.Foo", "StringTestMissingField.Foo", "Foo", "Foo", "fieldcontains")
}

func TestFieldExcludes(t *testing.T) {
	validate := New()
	type StringTest struct {
		Foo string `validate:"fieldexcludes=Bar"`
		Bar string
	}

	stringTest := &StringTest{
		Foo: "foobar",
		Bar: "bar",
	}

	errs := validate.Struct(stringTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "StringTest.Foo", "StringTest.Foo", "Foo", "Foo", "fieldexcludes")

	stringTest = &StringTest{
		Foo: "foo",
		Bar: "bar",
	}

	errs = validate.Struct(stringTest)
	Equal(t, errs, nil)

	errs = validate.VarWithValue("foo", "bar", "fieldexcludes")
	Equal(t, errs, nil)

	errs = validate.VarWithValue("bar", "foobarfoo", "fieldexcludes")
	Equal(t, errs, nil)

	errs = validate.VarWithValue("foobarfoo", "bar", "fieldexcludes")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "fieldexcludes")

	type StringTestMissingField struct {
		Foo string `validate:"fieldexcludes=Bar"`
	}

	stringTestMissingField := &StringTestMissingField{
		Foo: "foo",
	}

	errs = validate.Struct(stringTestMissingField)
	Equal(t, errs, nil)
}

func TestContainsAndExcludes(t *testing.T) {
	validate := New()
	type ImpossibleStringTest struct {
		Foo string `validate:"fieldcontains=Bar"`
		Bar string `validate:"fieldexcludes=Foo"`
	}

	impossibleStringTest := &ImpossibleStringTest{
		Foo: "foo",
		Bar: "bar",
	}

	errs := validate.Struct(impossibleStringTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "ImpossibleStringTest.Foo", "ImpossibleStringTest.Foo", "Foo", "Foo", "fieldcontains")

	impossibleStringTest = &ImpossibleStringTest{
		Foo: "bar",
		Bar: "foo",
	}

	errs = validate.Struct(impossibleStringTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "ImpossibleStringTest.Foo", "ImpossibleStringTest.Foo", "Foo", "Foo", "fieldcontains")
}

func TestLteField(t *testing.T) {
	var errs error
	validate := New()
	type TimeTest struct {
		Start *time.Time `validate:"required,lte,ltefield=End"`
		End   *time.Time `validate:"required,lte"`
	}

	now := time.Now()
	start := now.Add(time.Hour * 24 * -1 * 2)
	end := start.Add(time.Hour * 24)

	timeTest := &TimeTest{
		Start: &start,
		End:   &end,
	}

	errs = validate.Struct(timeTest)
	Equal(t, errs, nil)

	timeTest = &TimeTest{
		Start: &end,
		End:   &start,
	}

	errs = validate.Struct(timeTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeTest.Start", "TimeTest.Start", "Start", "Start", "ltefield")

	errs = validate.VarWithValue(&start, &end, "ltefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(&end, &start, "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	errs = validate.VarWithValue(&end, timeTest, "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	errs = validate.VarWithValue("tes", "test", "ltefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue("test", "test", "ltefield")
	Equal(t, errs, nil)

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "ltefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "ltefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	errs = validate.VarWithValue(time.Duration(0), -time.Minute, "omitempty,ltefield")
	Equal(t, errs, nil)

	// -- Validations for a struct with time.Duration type fields.

	type TimeDurationTest struct {
		First  time.Duration `validate:"ltefield=Second"`
		Second time.Duration
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "ltefield")

	type TimeDurationOmitemptyTest struct {
		First  time.Duration `validate:"omitempty,ltefield=Second"`
		Second time.Duration
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0), -time.Minute}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)

	// Tests for Ints types.

	type IntTest struct {
		Val1 int `validate:"required"`
		Val2 int `validate:"required,ltefield=Val1"`
	}

	intTest := &IntTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(intTest)
	Equal(t, errs, nil)

	intTest = &IntTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(intTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "IntTest.Val2", "IntTest.Val2", "Val2", "Val2", "ltefield")

	errs = validate.VarWithValue(int(1), int(5), "ltefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(int(5), int(1), "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	type UIntTest struct {
		Val1 uint `validate:"required"`
		Val2 uint `validate:"required,ltefield=Val1"`
	}

	uIntTest := &UIntTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(uIntTest)
	Equal(t, errs, nil)

	uIntTest = &UIntTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(uIntTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "UIntTest.Val2", "UIntTest.Val2", "Val2", "Val2", "ltefield")

	errs = validate.VarWithValue(uint(1), uint(5), "ltefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(uint(5), uint(1), "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	type FloatTest struct {
		Val1 float64 `validate:"required"`
		Val2 float64 `validate:"required,ltefield=Val1"`
	}

	floatTest := &FloatTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(floatTest)
	Equal(t, errs, nil)

	floatTest = &FloatTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(floatTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "FloatTest.Val2", "FloatTest.Val2", "Val2", "Val2", "ltefield")

	errs = validate.VarWithValue(float32(1), float32(5), "ltefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(float32(5), float32(1), "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	errs = validate.VarWithValue(nil, 5, "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	errs = validate.VarWithValue(1, "T", "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	errs = validate.VarWithValue(1, end, "ltefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "ltefield")

	type TimeTest2 struct {
		Start *time.Time `validate:"required"`
		End   *time.Time `validate:"required,ltefield=NonExistantField"`
	}

	timeTest2 := &TimeTest2{
		Start: &end,
		End:   &start,
	}

	errs = validate.Struct(timeTest2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeTest2.End", "TimeTest2.End", "End", "End", "ltefield")
}

func TestGteField(t *testing.T) {
	var errs error
	validate := New()
	type TimeTest struct {
		Start *time.Time `validate:"required,gte"`
		End   *time.Time `validate:"required,gte,gtefield=Start"`
	}

	now := time.Now()
	start := now.Add(time.Hour * 24)
	end := start.Add(time.Hour * 24)

	timeTest := &TimeTest{
		Start: &start,
		End:   &end,
	}

	errs = validate.Struct(timeTest)
	Equal(t, errs, nil)

	timeTest = &TimeTest{
		Start: &end,
		End:   &start,
	}

	errs = validate.Struct(timeTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeTest.End", "TimeTest.End", "End", "End", "gtefield")

	errs = validate.VarWithValue(&end, &start, "gtefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(&start, &end, "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	errs = validate.VarWithValue(&start, timeTest, "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	errs = validate.VarWithValue("test", "test", "gtefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue("test bigger", "test", "gtefield")
	Equal(t, errs, nil)

	// Tests for time.Duration type.

	// -- Validations for variables of time.Duration type.

	errs = validate.VarWithValue(time.Hour, time.Hour-time.Minute, "gtefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour, "gtefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(time.Hour, time.Hour+time.Minute, "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	errs = validate.VarWithValue(time.Duration(0), time.Hour, "omitempty,gtefield")
	Equal(t, errs, nil)

	// -- Validations for a struct with time.Duration type fields.

	type TimeDurationTest struct {
		First  time.Duration `validate:"gtefield=Second"`
		Second time.Duration
	}
	var timeDurationTest *TimeDurationTest

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour, time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.First", "TimeDurationTest.First", "First", "First", "gtefield")

	type TimeDurationOmitemptyTest struct {
		First  time.Duration `validate:"omitempty,gtefield=Second"`
		Second time.Duration
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0), time.Hour}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)

	// Tests for Ints types.

	type IntTest struct {
		Val1 int `validate:"required"`
		Val2 int `validate:"required,gtefield=Val1"`
	}

	intTest := &IntTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(intTest)
	Equal(t, errs, nil)

	intTest = &IntTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(intTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "IntTest.Val2", "IntTest.Val2", "Val2", "Val2", "gtefield")

	errs = validate.VarWithValue(int(5), int(1), "gtefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(int(1), int(5), "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	type UIntTest struct {
		Val1 uint `validate:"required"`
		Val2 uint `validate:"required,gtefield=Val1"`
	}

	uIntTest := &UIntTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(uIntTest)
	Equal(t, errs, nil)

	uIntTest = &UIntTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(uIntTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "UIntTest.Val2", "UIntTest.Val2", "Val2", "Val2", "gtefield")

	errs = validate.VarWithValue(uint(5), uint(1), "gtefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(uint(1), uint(5), "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	type FloatTest struct {
		Val1 float64 `validate:"required"`
		Val2 float64 `validate:"required,gtefield=Val1"`
	}

	floatTest := &FloatTest{
		Val1: 1,
		Val2: 5,
	}

	errs = validate.Struct(floatTest)
	Equal(t, errs, nil)

	floatTest = &FloatTest{
		Val1: 5,
		Val2: 1,
	}

	errs = validate.Struct(floatTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "FloatTest.Val2", "FloatTest.Val2", "Val2", "Val2", "gtefield")

	errs = validate.VarWithValue(float32(5), float32(1), "gtefield")
	Equal(t, errs, nil)

	errs = validate.VarWithValue(float32(1), float32(5), "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	errs = validate.VarWithValue(nil, 1, "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	errs = validate.VarWithValue(5, "T", "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	errs = validate.VarWithValue(5, start, "gtefield")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gtefield")

	type TimeTest2 struct {
		Start *time.Time `validate:"required"`
		End   *time.Time `validate:"required,gtefield=NonExistantField"`
	}

	timeTest2 := &TimeTest2{
		Start: &start,
		End:   &end,
	}

	errs = validate.Struct(timeTest2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeTest2.End", "TimeTest2.End", "End", "End", "gtefield")
}

func TestValidateByTagAndValue(t *testing.T) {
	validate := New()
	val := "test"
	field := "test"
	errs := validate.VarWithValue(val, field, "required")
	Equal(t, errs, nil)

	fn := func(fl FieldLevel) bool {
		return fl.Parent().String() == fl.Field().String()
	}
	errs = validate.RegisterValidation("isequaltestfunc", fn)
	Equal(t, errs, nil)

	errs = validate.VarWithValue(val, field, "isequaltestfunc")
	Equal(t, errs, nil)

	val = "unequal"
	errs = validate.VarWithValue(val, field, "isequaltestfunc")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "isequaltestfunc")
}

func TestAddFunctions(t *testing.T) {
	fn := func(fl FieldLevel) bool {
		return true
	}
	fnCtx := func(ctx context.Context, fl FieldLevel) bool {
		return true
	}
	validate := New()
	errs := validate.RegisterValidation("new", fn)
	Equal(t, errs, nil)

	errs = validate.RegisterValidation("", fn)
	NotEqual(t, errs, nil)

	errs = validate.RegisterValidation("new", nil)
	NotEqual(t, errs, nil)

	errs = validate.RegisterValidation("new", fn)
	Equal(t, errs, nil)

	errs = validate.RegisterValidationCtx("new", fnCtx)
	Equal(t, errs, nil)

	PanicMatches(t, func() { _ = validate.RegisterValidation("dive", fn) }, "Tag 'dive' either contains restricted characters or is the same as a restricted tag needed for normal operation")
}

func TestChangeTag(t *testing.T) {
	validate := New()
	validate.SetTagName("val")
	type Test struct {
		Name string `val:"len=4"`
	}

	s := &Test{
		Name: "TEST",
	}
	errs := validate.Struct(s)
	Equal(t, errs, nil)

	s.Name = ""
	errs = validate.Struct(s)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.Name", "Test.Name", "Name", "Name", "len")
}

func TestUnexposedStruct(t *testing.T) {
	validate := New()
	type Test struct {
		Name      string
		unexposed struct {
			A string `validate:"required"`
		}
	}

	s := &Test{
		Name: "TEST",
	}
	Equal(t, s.unexposed.A, "")

	errs := validate.Struct(s)
	Equal(t, errs, nil)
}

func TestBadParams(t *testing.T) {
	validate := New()
	i := 1
	errs := validate.Var(i, "-")
	Equal(t, errs, nil)

	PanicMatches(t, func() { _ = validate.Var(i, "len=a") }, "strconv.ParseInt: parsing \"a\": invalid syntax")
	PanicMatches(t, func() { _ = validate.Var(i, "len=a") }, "strconv.ParseInt: parsing \"a\": invalid syntax")

	var ui uint = 1
	PanicMatches(t, func() { _ = validate.Var(ui, "len=a") }, "strconv.ParseUint: parsing \"a\": invalid syntax")

	f := 1.23
	PanicMatches(t, func() { _ = validate.Var(f, "len=a") }, "strconv.ParseFloat: parsing \"a\": invalid syntax")
}

func TestLength(t *testing.T) {
	validate := New()
	i := true
	PanicMatches(t, func() { _ = validate.Var(i, "len") }, "Bad field type bool")
}

func TestIsGt(t *testing.T) {
	var errs error
	validate := New()
	myMap := map[string]string{}
	errs = validate.Var(myMap, "gt=0")
	NotEqual(t, errs, nil)

	f := 1.23
	errs = validate.Var(f, "gt=5")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gt")

	var ui uint = 5
	errs = validate.Var(ui, "gt=10")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gt")

	i := true
	PanicMatches(t, func() { _ = validate.Var(i, "gt") }, "Bad field type bool")

	tm := time.Now().UTC()
	tm = tm.Add(time.Hour * 24)

	errs = validate.Var(tm, "gt")
	Equal(t, errs, nil)

	t2 := time.Now().UTC().Add(-time.Hour)

	errs = validate.Var(t2, "gt")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gt")

	type Test struct {
		Now *time.Time `validate:"gt"`
	}
	s := &Test{
		Now: &tm,
	}
	errs = validate.Struct(s)
	Equal(t, errs, nil)

	s = &Test{
		Now: &t2,
	}

	errs = validate.Struct(s)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.Now", "Test.Now", "Now", "Now", "gt")

	// Tests for time.Duration type.

	// -- Validations for a variable of time.Duration type.

	errs = validate.Var(time.Hour, "gt=59m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-time.Minute, "gt=59m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gt")

	errs = validate.Var(time.Hour-2*time.Minute, "gt=59m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gt")

	errs = validate.Var(time.Duration(0), "omitempty,gt=59m")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.

	type TimeDurationTest struct {
		Duration time.Duration `validate:"gt=59m"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "gt")

	timeDurationTest = &TimeDurationTest{time.Hour - 2*time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "gt")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,gt=59m"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestIsGte(t *testing.T) {
	var errs error
	validate := New()
	i := true
	PanicMatches(t, func() { _ = validate.Var(i, "gte") }, "Bad field type bool")

	t1 := time.Now().UTC()
	t1 = t1.Add(time.Hour * 24)

	errs = validate.Var(t1, "gte")
	Equal(t, errs, nil)

	t2 := time.Now().UTC().Add(-time.Hour)

	errs = validate.Var(t2, "gte")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gte")

	type Test struct {
		Now *time.Time `validate:"gte"`
	}
	s := &Test{
		Now: &t1,
	}

	errs = validate.Struct(s)
	Equal(t, errs, nil)

	s = &Test{
		Now: &t2,
	}

	errs = validate.Struct(s)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.Now", "Test.Now", "Now", "Now", "gte")

	// Tests for time.Duration type.

	// -- Validations for a variable of time.Duration type.

	errs = validate.Var(time.Hour, "gte=59m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-time.Minute, "gte=59m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-2*time.Minute, "gte=59m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "gte")

	errs = validate.Var(time.Duration(0), "omitempty,gte=59m")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.

	type TimeDurationTest struct {
		Duration time.Duration `validate:"gte=59m"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - 2*time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "gte")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,gte=59m"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestMinValidation(t *testing.T) {
	var errs error
	validate := New()
	// Tests for time.Duration type.

	// -- Validations for a variable of time.Duration type.

	errs = validate.Var(time.Hour, "min=59m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-time.Minute, "min=59m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-2*time.Minute, "min=59m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "min")

	errs = validate.Var(time.Duration(0), "omitempty,min=59m")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.

	type TimeDurationTest struct {
		Duration time.Duration `validate:"min=59m"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - 2*time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "min")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,min=59m"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestMaxValidation(t *testing.T) {
	var errs error
	validate := New()
	// Tests for time.Duration type.
	// -- Validations for a variable of time.Duration type.

	errs = validate.Var(time.Hour, "max=1h1m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour+time.Minute, "max=1h1m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour+2*time.Minute, "max=1h1m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "max")

	errs = validate.Var(time.Duration(0), "omitempty,max=-1s")
	Equal(t, errs, nil)
	// -- Validations for a struct with a time.Duration type field.
	type TimeDurationTest struct {
		Duration time.Duration `validate:"max=1h1m"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour + 2*time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "max")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,max=-1s"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestMinMaxValidation(t *testing.T) {
	var errs error
	validate := New()
	// Tests for time.Duration type.
	// -- Validations for a variable of time.Duration type.
	errs = validate.Var(time.Hour, "min=59m,max=1h1m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-time.Minute, "min=59m,max=1h1m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour+time.Minute, "min=59m,max=1h1m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-2*time.Minute, "min=59m,max=1h1m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "min")

	errs = validate.Var(time.Hour+2*time.Minute, "min=59m,max=1h1m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "max")

	errs = validate.Var(time.Duration(0), "omitempty,min=59m,max=1h1m")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.
	type TimeDurationTest struct {
		Duration time.Duration `validate:"min=59m,max=1h1m"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - 2*time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "min")

	timeDurationTest = &TimeDurationTest{time.Hour + 2*time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "max")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,min=59m,max=1h1m"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestLenValidation(t *testing.T) {
	var errs error
	validate := New()
	// Tests for time.Duration type.
	// -- Validations for a variable of time.Duration type.
	errs = validate.Var(time.Hour, "len=1h")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour-time.Minute, "len=1h")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "len")

	errs = validate.Var(time.Hour+time.Minute, "len=1h")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "len")

	errs = validate.Var(time.Duration(0), "omitempty,len=1h")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.
	type TimeDurationTest struct {
		Duration time.Duration `validate:"len=1h"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour - time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "len")

	timeDurationTest = &TimeDurationTest{time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "len")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,len=1h"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestIsLt(t *testing.T) {
	var errs error
	validate := New()
	myMap := map[string]string{}
	errs = validate.Var(myMap, "lt=0")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "lt")

	f := 1.23
	errs = validate.Var(f, "lt=0")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "lt")

	var ui uint = 5
	errs = validate.Var(ui, "lt=0")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "lt")

	i := true
	PanicMatches(t, func() { _ = validate.Var(i, "lt") }, "Bad field type bool")

	t1 := time.Now().UTC().Add(-time.Hour)

	errs = validate.Var(t1, "lt")
	Equal(t, errs, nil)

	t2 := time.Now().UTC()
	t2 = t2.Add(time.Hour * 24)

	errs = validate.Var(t2, "lt")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "lt")

	type Test struct {
		Now *time.Time `validate:"lt"`
	}

	s := &Test{
		Now: &t1,
	}
	errs = validate.Struct(s)
	Equal(t, errs, nil)

	s = &Test{
		Now: &t2,
	}

	errs = validate.Struct(s)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.Now", "Test.Now", "Now", "Now", "lt")

	// Tests for time.Duration type.
	// -- Validations for a variable of time.Duration type.
	errs = validate.Var(time.Hour, "lt=1h1m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour+time.Minute, "lt=1h1m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "lt")

	errs = validate.Var(time.Hour+2*time.Minute, "lt=1h1m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "lt")

	errs = validate.Var(time.Duration(0), "omitempty,lt=0")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.
	type TimeDurationTest struct {
		Duration time.Duration `validate:"lt=1h1m"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "lt")

	timeDurationTest = &TimeDurationTest{time.Hour + 2*time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "lt")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,lt=0"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestIsLte(t *testing.T) {
	var errs error
	validate := New()
	i := true
	PanicMatches(t, func() { _ = validate.Var(i, "lte") }, "Bad field type bool")

	t1 := time.Now().UTC().Add(-time.Hour)
	errs = validate.Var(t1, "lte")
	Equal(t, errs, nil)

	t2 := time.Now().UTC()
	t2 = t2.Add(time.Hour * 24)

	errs = validate.Var(t2, "lte")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "lte")

	type Test struct {
		Now *time.Time `validate:"lte"`
	}

	s := &Test{
		Now: &t1,
	}

	errs = validate.Struct(s)
	Equal(t, errs, nil)

	s = &Test{
		Now: &t2,
	}

	errs = validate.Struct(s)
	NotEqual(t, errs, nil)

	// Tests for time.Duration type.
	// -- Validations for a variable of time.Duration type.
	errs = validate.Var(time.Hour, "lte=1h1m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour+time.Minute, "lte=1h1m")
	Equal(t, errs, nil)

	errs = validate.Var(time.Hour+2*time.Minute, "lte=1h1m")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "lte")

	errs = validate.Var(time.Duration(0), "omitempty,lte=-1s")
	Equal(t, errs, nil)

	// -- Validations for a struct with a time.Duration type field.
	type TimeDurationTest struct {
		Duration time.Duration `validate:"lte=1h1m"`
	}

	var timeDurationTest *TimeDurationTest
	timeDurationTest = &TimeDurationTest{time.Hour}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour + time.Minute}
	errs = validate.Struct(timeDurationTest)
	Equal(t, errs, nil)

	timeDurationTest = &TimeDurationTest{time.Hour + 2*time.Minute}
	errs = validate.Struct(timeDurationTest)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TimeDurationTest.Duration", "TimeDurationTest.Duration", "Duration", "Duration", "lte")

	type TimeDurationOmitemptyTest struct {
		Duration time.Duration `validate:"omitempty,lte=-1s"`
	}

	timeDurationOmitemptyTest := &TimeDurationOmitemptyTest{time.Duration(0)}
	errs = validate.Struct(timeDurationOmitemptyTest)
	Equal(t, errs, nil)
}

func TestHsla(t *testing.T) {
	validate := New()
	s := "hsla(360,100%,100%,1)"
	errs := validate.Var(s, "hsla")
	Equal(t, errs, nil)

	s = "hsla(360,100%,100%,0.5)"
	errs = validate.Var(s, "hsla")
	Equal(t, errs, nil)

	s = "hsla(0,0%,0%, 0)"
	errs = validate.Var(s, "hsla")
	Equal(t, errs, nil)

	s = "hsl(361,100%,50%,1)"
	errs = validate.Var(s, "hsla")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsla")

	s = "hsl(361,100%,50%)"
	errs = validate.Var(s, "hsla")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsla")

	s = "hsla(361,100%,50%)"
	errs = validate.Var(s, "hsla")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsla")

	s = "hsla(360,101%,50%)"
	errs = validate.Var(s, "hsla")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsla")

	s = "hsla(360,100%,101%)"
	errs = validate.Var(s, "hsla")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsla")

	i := 1
	errs = validate.Var(i, "hsla")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsla")
}

func TestHsl(t *testing.T) {
	validate := New()
	s := "hsl(360,100%,50%)"
	errs := validate.Var(s, "hsl")
	Equal(t, errs, nil)

	s = "hsl(0,0%,0%)"
	errs = validate.Var(s, "hsl")
	Equal(t, errs, nil)

	s = "hsl(361,100%,50%)"
	errs = validate.Var(s, "hsl")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsl")

	s = "hsl(361,101%,50%)"
	errs = validate.Var(s, "hsl")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsl")

	s = "hsl(361,100%,101%)"
	errs = validate.Var(s, "hsl")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsl")

	s = "hsl(-10,100%,100%)"
	errs = validate.Var(s, "hsl")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsl")

	i := 1
	errs = validate.Var(i, "hsl")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hsl")
}

func TestRgba(t *testing.T) {
	validate := New()
	s := "rgba(0,31,255,0.5)"
	errs := validate.Var(s, "rgba")
	Equal(t, errs, nil)

	s = "rgba(0,31,255,0.12)"
	errs = validate.Var(s, "rgba")
	Equal(t, errs, nil)

	s = "rgba(12%,55%,100%,0.12)"
	errs = validate.Var(s, "rgba")
	Equal(t, errs, nil)

	s = "rgba( 0,  31, 255, 0.5)"
	errs = validate.Var(s, "rgba")
	Equal(t, errs, nil)

	s = "rgba(12%,55,100%,0.12)"
	errs = validate.Var(s, "rgba")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgba")

	s = "rgb(0,  31, 255)"
	errs = validate.Var(s, "rgba")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgba")

	s = "rgb(1,349,275,0.5)"
	errs = validate.Var(s, "rgba")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgba")

	s = "rgb(01,31,255,0.5)"
	errs = validate.Var(s, "rgba")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgba")

	i := 1
	errs = validate.Var(i, "rgba")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgba")
}

func TestRgb(t *testing.T) {
	validate := New()
	s := "rgb(0,31,255)"
	errs := validate.Var(s, "rgb")
	Equal(t, errs, nil)

	s = "rgb(0,  31, 255)"
	errs = validate.Var(s, "rgb")
	Equal(t, errs, nil)

	s = "rgb(10%,  50%, 100%)"
	errs = validate.Var(s, "rgb")
	Equal(t, errs, nil)

	s = "rgb(10%,  50%, 55)"
	errs = validate.Var(s, "rgb")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgb")

	s = "rgb(1,349,275)"
	errs = validate.Var(s, "rgb")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgb")

	s = "rgb(01,31,255)"
	errs = validate.Var(s, "rgb")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgb")

	s = "rgba(0,31,255)"
	errs = validate.Var(s, "rgb")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgb")

	i := 1
	errs = validate.Var(i, "rgb")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "rgb")
}

func TestEmail(t *testing.T) {
	validate := New()
	s := "test@mail.com"
	errs := validate.Var(s, "email")
	Equal(t, errs, nil)

	s = "Dörte@Sörensen.example.com"
	errs = validate.Var(s, "email")
	Equal(t, errs, nil)

	s = "θσερ@εχαμπλε.ψομ"
	errs = validate.Var(s, "email")
	Equal(t, errs, nil)

	s = "юзер@екзампл.ком"
	errs = validate.Var(s, "email")
	Equal(t, errs, nil)

	s = "उपयोगकर्ता@उदाहरण.कॉम"
	errs = validate.Var(s, "email")
	Equal(t, errs, nil)

	s = "用户@例子.广告"
	errs = validate.Var(s, "email")
	Equal(t, errs, nil)

	s = "mail@domain_with_underscores.org"
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")

	s = "mail@dotaftercom.com."
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")

	s = "mail@dotaftercom.co.uk."
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")

	s = "Foo Bar <foobar@example.com>"
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)

	s = ""
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")

	s = "test@email"
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")

	s = "test@email."
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")

	s = "@email.com"
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")

	s = `"test test"@email.com`
	errs = validate.Var(s, "email")
	Equal(t, errs, nil)

	s = `"@email.com`
	errs = validate.Var(s, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")

	i := true
	errs = validate.Var(i, "email")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "email")
}

func TestHexColor(t *testing.T) {
	validate := New()
	s := "#fff"
	errs := validate.Var(s, "hexcolor")
	Equal(t, errs, nil)

	s = "#c2c2c2"
	errs = validate.Var(s, "hexcolor")
	Equal(t, errs, nil)

	s = "fff"
	errs = validate.Var(s, "hexcolor")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hexcolor")

	s = "fffFF"
	errs = validate.Var(s, "hexcolor")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hexcolor")

	i := true
	errs = validate.Var(i, "hexcolor")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hexcolor")
}

func TestHexadecimal(t *testing.T) {
	validate := New()
	s := "ff0044"
	errs := validate.Var(s, "hexadecimal")
	Equal(t, errs, nil)

	s = "0xff0044"
	errs = validate.Var(s, "hexadecimal")
	Equal(t, errs, nil)

	s = "0Xff0044"
	errs = validate.Var(s, "hexadecimal")
	Equal(t, errs, nil)

	s = "abcdefg"
	errs = validate.Var(s, "hexadecimal")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hexadecimal")

	i := true
	errs = validate.Var(i, "hexadecimal")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "hexadecimal")
}

func TestNumber(t *testing.T) {
	validate := New()
	s := "1"
	errs := validate.Var(s, "number")
	Equal(t, errs, nil)

	s = "+1"
	errs = validate.Var(s, "number")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "number")

	s = "-1"
	errs = validate.Var(s, "number")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "number")

	s = "1.12"
	errs = validate.Var(s, "number")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "number")

	s = "+1.12"
	errs = validate.Var(s, "number")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "number")

	s = "-1.12"
	errs = validate.Var(s, "number")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "number")

	s = "1."
	errs = validate.Var(s, "number")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "number")

	s = "1.o"
	errs = validate.Var(s, "number")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "number")

	i := 1
	errs = validate.Var(i, "number")
	Equal(t, errs, nil)
}

func TestNumeric(t *testing.T) {
	validate := New()
	s := "1"
	errs := validate.Var(s, "numeric")
	Equal(t, errs, nil)

	s = "+1"
	errs = validate.Var(s, "numeric")
	Equal(t, errs, nil)

	s = "-1"
	errs = validate.Var(s, "numeric")
	Equal(t, errs, nil)

	s = "1.12"
	errs = validate.Var(s, "numeric")
	Equal(t, errs, nil)

	s = "+1.12"
	errs = validate.Var(s, "numeric")
	Equal(t, errs, nil)

	s = "-1.12"
	errs = validate.Var(s, "numeric")
	Equal(t, errs, nil)

	s = "1."
	errs = validate.Var(s, "numeric")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "numeric")

	s = "1.o"
	errs = validate.Var(s, "numeric")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "numeric")

	i := 1
	errs = validate.Var(i, "numeric")
	Equal(t, errs, nil)
}

func TestBoolean(t *testing.T) {
	validate := New()
	b := true
	errs := validate.Var(b, "boolean")
	Equal(t, errs, nil)

	b = false
	errs = validate.Var(b, "boolean")
	Equal(t, errs, nil)

	s := "true"
	errs = validate.Var(s, "boolean")
	Equal(t, errs, nil)

	s = "false"
	errs = validate.Var(s, "boolean")
	Equal(t, errs, nil)

	s = "0"
	errs = validate.Var(s, "boolean")
	Equal(t, errs, nil)

	s = "1"
	errs = validate.Var(s, "boolean")
	Equal(t, errs, nil)

	s = "xyz"
	errs = validate.Var(s, "boolean")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "boolean")

	s = "1."
	errs = validate.Var(s, "boolean")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "boolean")
}

func TestAlphaNumeric(t *testing.T) {
	validate := New()
	s := "abcd123"
	errs := validate.Var(s, "alphanum")
	Equal(t, errs, nil)

	s = "abc!23"
	errs = validate.Var(s, "alphanum")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "alphanum")

	errs = validate.Var(1, "alphanum")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "alphanum")
}

func TestAlpha(t *testing.T) {
	validate := New()
	s := "abcd"
	errs := validate.Var(s, "alpha")
	Equal(t, errs, nil)

	s = "abc®"
	errs = validate.Var(s, "alpha")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "alpha")

	s = "abc÷"
	errs = validate.Var(s, "alpha")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "alpha")

	s = "abc1"
	errs = validate.Var(s, "alpha")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "alpha")

	s = "this is a test string"
	errs = validate.Var(s, "alpha")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "alpha")

	errs = validate.Var(1, "alpha")
	NotEqual(t, errs, nil)
	AssertError(t, errs, "", "", "", "", "alpha")
}

func TestStructInt32Validation(t *testing.T) {
	type TestInt32 struct {
		Required  int `validate:"required"`
		Len       int `validate:"len=10"`
		Min       int `validate:"min=1"`
		Max       int `validate:"max=10"`
		MinMax    int `validate:"min=1,max=10"`
		Lt        int `validate:"lt=10"`
		Lte       int `validate:"lte=10"`
		Gt        int `validate:"gt=10"`
		Gte       int `validate:"gte=10"`
		OmitEmpty int `validate:"omitempty,min=1,max=10"`
	}

	tSuccess := &TestInt32{
		Required:  1,
		Len:       10,
		Min:       1,
		Max:       10,
		MinMax:    5,
		Lt:        9,
		Lte:       10,
		Gt:        11,
		Gte:       10,
		OmitEmpty: 0,
	}
	validate := New()
	errs := validate.Struct(tSuccess)
	Equal(t, errs, nil)

	tFail := &TestInt32{
		Required:  0,
		Len:       11,
		Min:       -1,
		Max:       11,
		MinMax:    -1,
		Lt:        10,
		Lte:       11,
		Gt:        10,
		Gte:       9,
		OmitEmpty: 11,
	}
	errs = validate.Struct(tFail)

	// Assert Top Level
	NotEqual(t, errs, nil)
	Equal(t, len(errs.(ValidationErrors)), 10)

	// Assert Fields
	AssertError(t, errs, "TestInt32.Required", "TestInt32.Required", "Required", "Required", "required")
	AssertError(t, errs, "TestInt32.Len", "TestInt32.Len", "Len", "Len", "len")
	AssertError(t, errs, "TestInt32.Min", "TestInt32.Min", "Min", "Min", "min")
	AssertError(t, errs, "TestInt32.Max", "TestInt32.Max", "Max", "Max", "max")
	AssertError(t, errs, "TestInt32.MinMax", "TestInt32.MinMax", "MinMax", "MinMax", "min")
	AssertError(t, errs, "TestInt32.Lt", "TestInt32.Lt", "Lt", "Lt", "lt")
	AssertError(t, errs, "TestInt32.Lte", "TestInt32.Lte", "Lte", "Lte", "lte")
	AssertError(t, errs, "TestInt32.Gt", "TestInt32.Gt", "Gt", "Gt", "gt")
	AssertError(t, errs, "TestInt32.Gte", "TestInt32.Gte", "Gte", "Gte", "gte")
	AssertError(t, errs, "TestInt32.OmitEmpty", "TestInt32.OmitEmpty", "OmitEmpty", "OmitEmpty", "max")
}

func TestMultipleRecursiveExtractStructCache(t *testing.T) {
	validate := New()
	type Recursive struct {
		Field *string `validate:"required,len=5,ne=string"`
	}

	var test Recursive
	current := reflect.ValueOf(test)
	name := "Recursive"
	proceed := make(chan struct{})
	sc := validate.extractStructCache(current, name)
	ptr := fmt.Sprintf("%p", sc)
	for i := 0; i < 100; i++ {
		go func() {
			<-proceed
			sc := validate.extractStructCache(current, name)
			Equal(t, ptr, fmt.Sprintf("%p", sc))
		}()
	}

	close(proceed)
}

func TestPointerAndOmitEmpty(t *testing.T) {
	validate := New()
	type Test struct {
		MyInt *int `validate:"omitempty,gte=2,lte=255"`
	}

	var val1 int
	val2 := 256
	t1 := Test{MyInt: &val1} // This should fail validation on gte because value is 0
	t2 := Test{MyInt: &val2} // This should fail validate on lte because value is 256
	t3 := Test{MyInt: nil}   // This should succeed validation because pointer is nil

	errs := validate.Struct(t1)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.MyInt", "Test.MyInt", "MyInt", "MyInt", "gte")

	errs = validate.Struct(t2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "Test.MyInt", "Test.MyInt", "MyInt", "MyInt", "lte")

	errs = validate.Struct(t3)
	Equal(t, errs, nil)

	type TestIface struct {
		MyInt interface{} `validate:"omitempty,gte=2,lte=255"`
	}

	ti1 := TestIface{MyInt: &val1} // This should fail validation on gte because value is 0
	ti2 := TestIface{MyInt: &val2} // This should fail validate on lte because value is 256
	ti3 := TestIface{MyInt: nil}   // This should succeed validation because pointer is nil
	errs = validate.Struct(ti1)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TestIface.MyInt", "TestIface.MyInt", "MyInt", "MyInt", "gte")

	errs = validate.Struct(ti2)
	NotEqual(t, errs, nil)
	AssertError(t, errs, "TestIface.MyInt", "TestIface.MyInt", "MyInt", "MyInt", "lte")

	errs = validate.Struct(ti3)
	Equal(t, errs, nil)
}

func TestRequired(t *testing.T) {
	validate := New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	type Test struct {
		Value interface{} `validate:"required"`
	}

	var test Test
	err := validate.Struct(test)
	NotEqual(t, err, nil)
	AssertError(t, err.(ValidationErrors), "Test.Value", "Test.Value", "Value", "Value", "required")
}

func TestBoolEqual(t *testing.T) {
	validate := New()
	type Test struct {
		Value bool `validate:"eq=true"`
	}

	var test Test
	err := validate.Struct(test)
	NotEqual(t, err, nil)
	AssertError(t, err.(ValidationErrors), "Test.Value", "Test.Value", "Value", "Value", "eq")

	test.Value = true
	err = validate.Struct(test)
	Equal(t, err, nil)
}

func TestRequiredPtr(t *testing.T) {
	type Test struct {
		Bool *bool `validate:"required"`
	}

	var f bool
	validate := New()
	test := Test{
		Bool: &f,
	}

	err := validate.Struct(test)
	Equal(t, err, nil)

	tr := true
	test.Bool = &tr
	err = validate.Struct(test)
	Equal(t, err, nil)

	test.Bool = nil

	err = validate.Struct(test)
	NotEqual(t, err, nil)

	errs, ok := err.(ValidationErrors)
	Equal(t, ok, true)
	Equal(t, len(errs), 1)
	AssertError(t, errs, "Test.Bool", "Test.Bool", "Bool", "Bool", "required")

	type Test2 struct {
		Bool bool `validate:"required"`
	}

	var test2 Test2
	err = validate.Struct(test2)
	NotEqual(t, err, nil)

	errs, ok = err.(ValidationErrors)
	Equal(t, ok, true)
	Equal(t, len(errs), 1)
	AssertError(t, errs, "Test2.Bool", "Test2.Bool", "Bool", "Bool", "required")

	test2.Bool = true
	err = validate.Struct(test2)
	Equal(t, err, nil)

	type Test3 struct {
		Arr []string `validate:"required"`
	}

	var test3 Test3
	err = validate.Struct(test3)
	NotEqual(t, err, nil)

	errs, ok = err.(ValidationErrors)
	Equal(t, ok, true)
	Equal(t, len(errs), 1)
	AssertError(t, errs, "Test3.Arr", "Test3.Arr", "Arr", "Arr", "required")

	test3.Arr = make([]string, 0)
	err = validate.Struct(test3)
	Equal(t, err, nil)

	type Test4 struct {
		Arr *[]string `validate:"required"` // I know I know pointer to array, just making sure validation works as expected...
	}

	var test4 Test4
	err = validate.Struct(test4)
	NotEqual(t, err, nil)

	errs, ok = err.(ValidationErrors)
	Equal(t, ok, true)
	Equal(t, len(errs), 1)
	AssertError(t, errs, "Test4.Arr", "Test4.Arr", "Arr", "Arr", "required")

	arr := make([]string, 0)
	test4.Arr = &arr
	err = validate.Struct(test4)
	Equal(t, err, nil)
}

func TestArrayStructNamespace(t *testing.T) {
	validate := New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "-" {
			return name
		}

		return ""
	})

	type child struct {
		Name string `json:"name" validate:"required"`
	}

	var input struct {
		Children []child `json:"children" validate:"required,gt=0,dive"`
	}
	input.Children = []child{{"ok"}, {""}}
	errs := validate.Struct(input)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "children[1].name", "Children[1].Name", "name", "Name", "required")
}

func TestMapStructNamespace(t *testing.T) {
	validate := New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "-" {
			return name
		}

		return ""
	})

	type child struct {
		Name string `json:"name" validate:"required"`
	}

	var input struct {
		Children map[int]child `json:"children" validate:"required,gt=0,dive"`
	}
	input.Children = map[int]child{
		0: {Name: "ok"},
		1: {Name: ""},
	}

	errs := validate.Struct(input)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "children[1].name", "Children[1].Name", "name", "Name", "required")
}

func TestFieldLevelName(t *testing.T) {
	type Test struct {
		String string            `validate:"custom1"      json:"json1"`
		Array  []string          `validate:"dive,custom2" json:"json2"`
		Map    map[string]string `validate:"dive,custom3" json:"json3"`
		Array2 []string          `validate:"custom4"      json:"json4"`
		Map2   map[string]string `validate:"custom5"      json:"json5"`
	}

	var res1, res2, res3, res4, res5, alt1, alt2, alt3, alt4, alt5 string
	validate := New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})
	err := validate.RegisterValidation("custom1", func(fl FieldLevel) bool {
		res1 = fl.FieldName()
		alt1 = fl.StructFieldName()
		return true
	})
	Equal(t, err, nil)

	err = validate.RegisterValidation("custom2", func(fl FieldLevel) bool {
		res2 = fl.FieldName()
		alt2 = fl.StructFieldName()
		return true
	})
	Equal(t, err, nil)

	err = validate.RegisterValidation("custom3", func(fl FieldLevel) bool {
		res3 = fl.FieldName()
		alt3 = fl.StructFieldName()
		return true
	})
	Equal(t, err, nil)

	err = validate.RegisterValidation("custom4", func(fl FieldLevel) bool {
		res4 = fl.FieldName()
		alt4 = fl.StructFieldName()
		return true
	})
	Equal(t, err, nil)

	err = validate.RegisterValidation("custom5", func(fl FieldLevel) bool {
		res5 = fl.FieldName()
		alt5 = fl.StructFieldName()
		return true
	})
	Equal(t, err, nil)

	test := Test{
		String: "test",
		Array:  []string{"1"},
		Map:    map[string]string{"test": "test"},
	}

	errs := validate.Struct(test)
	Equal(t, errs, nil)
	Equal(t, res1, "json1")
	Equal(t, alt1, "String")
	Equal(t, res2, "json2[0]")
	Equal(t, alt2, "Array[0]")
	Equal(t, res3, "json3[test]")
	Equal(t, alt3, "Map[test]")
	Equal(t, res4, "json4")
	Equal(t, alt4, "Array2")
	Equal(t, res5, "json5")
	Equal(t, alt5, "Map2")
}

func TestValidateStructRegisterCtx(t *testing.T) {
	var ctxVal string
	fnCtx := func(ctx context.Context, fl FieldLevel) bool {
		ctxVal = ctx.Value(&ctxVal).(string)
		return true
	}

	var ctxSlVal string
	slFn := func(ctx context.Context, sl StructLevel) {
		ctxSlVal = ctx.Value(&ctxSlVal).(string)
	}

	type Test struct {
		Field string `validate:"val"`
	}

	var tst Test
	validate := New()
	err := validate.RegisterValidationCtx("val", fnCtx)
	Equal(t, err, nil)

	validate.RegisterStructValidationCtx(slFn, Test{})

	ctx := context.WithValue(context.Background(), &ctxVal, "testval")
	ctx = context.WithValue(ctx, &ctxSlVal, "slVal")
	errs := validate.StructCtx(ctx, tst)
	Equal(t, errs, nil)
	Equal(t, ctxVal, "testval")
	Equal(t, ctxSlVal, "slVal")
}

func TestIsDefault(t *testing.T) {
	validate := New()
	type Inner struct {
		String string `validate:"isdefault"`
	}

	type Test struct {
		String string `validate:"isdefault"`
		Inner  *Inner `validate:"isdefault"`
	}

	var tt Test
	errs := validate.Struct(tt)
	Equal(t, errs, nil)

	tt.Inner = &Inner{String: ""}
	errs = validate.Struct(tt)
	NotEqual(t, errs, nil)

	fe := errs.(ValidationErrors)[0]
	Equal(t, fe.Field(), "Inner")
	Equal(t, fe.Namespace(), "Test.Inner")
	Equal(t, fe.Tag(), "isdefault")

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	type Inner2 struct {
		String string `validate:"isdefault"`
	}

	type Test2 struct {
		Inner Inner2 `validate:"isdefault" json:"inner"`
	}

	var t2 Test2
	errs = validate.Struct(t2)
	Equal(t, errs, nil)

	t2.Inner.String = "Changed"
	errs = validate.Struct(t2)
	NotEqual(t, errs, nil)

	fe = errs.(ValidationErrors)[0]
	Equal(t, fe.Field(), "inner")
	Equal(t, fe.Namespace(), "Test2.inner")
	Equal(t, fe.Tag(), "isdefault")
}

func TestKeys(t *testing.T) {
	type Test struct {
		Test1 map[string]string `validate:"gt=0,dive,keys,eq=testkey,endkeys,eq=testval" json:"test1"`
		Test2 map[int]int       `validate:"gt=0,dive,keys,eq=3,endkeys,eq=4"             json:"test2"`
		Test3 map[int]int       `validate:"gt=0,dive,keys,eq=3,endkeys"                  json:"test3"`
	}

	var tst Test
	validate := New()
	err := validate.Struct(tst)
	NotEqual(t, err, nil)
	Equal(t, len(err.(ValidationErrors)), 3)
	AssertError(t, err.(ValidationErrors), "Test.Test1", "Test.Test1", "Test1", "Test1", "gt")
	AssertError(t, err.(ValidationErrors), "Test.Test2", "Test.Test2", "Test2", "Test2", "gt")
	AssertError(t, err.(ValidationErrors), "Test.Test3", "Test.Test3", "Test3", "Test3", "gt")

	tst.Test1 = map[string]string{
		"testkey": "testval",
	}

	tst.Test2 = map[int]int{
		3: 4,
	}

	tst.Test3 = map[int]int{
		3: 4,
	}

	err = validate.Struct(tst)
	Equal(t, err, nil)

	tst.Test1["badtestkey"] = "badtestvalue"
	tst.Test2[10] = 11

	err = validate.Struct(tst)
	NotEqual(t, err, nil)

	errs := err.(ValidationErrors)
	Equal(t, len(errs), 4)

	AssertDeepError(t, errs, "Test.Test1[badtestkey]", "Test.Test1[badtestkey]", "Test1[badtestkey]", "Test1[badtestkey]", "eq", "eq")
	AssertDeepError(t, errs, "Test.Test1[badtestkey]", "Test.Test1[badtestkey]", "Test1[badtestkey]", "Test1[badtestkey]", "eq", "eq")
	AssertDeepError(t, errs, "Test.Test2[10]", "Test.Test2[10]", "Test2[10]", "Test2[10]", "eq", "eq")
	AssertDeepError(t, errs, "Test.Test2[10]", "Test.Test2[10]", "Test2[10]", "Test2[10]", "eq", "eq")

	type Test2 struct {
		NestedKeys map[[1]string]string `validate:"gt=0,dive,keys,dive,eq=innertestkey,endkeys,eq=outertestval"`
	}

	var tst2 Test2
	err = validate.Struct(tst2)
	NotEqual(t, err, nil)
	Equal(t, len(err.(ValidationErrors)), 1)
	AssertError(t, err.(ValidationErrors), "Test2.NestedKeys", "Test2.NestedKeys", "NestedKeys", "NestedKeys", "gt")

	tst2.NestedKeys = map[[1]string]string{
		{"innertestkey"}: "outertestval",
	}

	err = validate.Struct(tst2)
	Equal(t, err, nil)

	tst2.NestedKeys[[1]string{"badtestkey"}] = "badtestvalue"
	err = validate.Struct(tst2)
	NotEqual(t, err, nil)

	errs = err.(ValidationErrors)
	Equal(t, len(errs), 2)
	AssertDeepError(t, errs, "Test2.NestedKeys[[badtestkey]][0]", "Test2.NestedKeys[[badtestkey]][0]", "NestedKeys[[badtestkey]][0]", "NestedKeys[[badtestkey]][0]", "eq", "eq")
	AssertDeepError(t, errs, "Test2.NestedKeys[[badtestkey]]", "Test2.NestedKeys[[badtestkey]]", "NestedKeys[[badtestkey]]", "NestedKeys[[badtestkey]]", "eq", "eq")

	// test bad tag definitions

	PanicMatches(t, func() { _ = validate.Var(map[string]string{"key": "val"}, "endkeys,dive,eq=val") }, "'endkeys' tag encountered without a corresponding 'keys' tag")
	PanicMatches(t, func() { _ = validate.Var(1, "keys,eq=1,endkeys") }, "'keys' tag must be immediately preceded by the 'dive' tag")

	// test custom tag name
	validate = New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "-" {
			return name
		}

		return ""
	})

	err = validate.Struct(tst)
	NotEqual(t, err, nil)

	errs = err.(ValidationErrors)
	Equal(t, len(errs), 4)

	AssertDeepError(t, errs, "Test.test1[badtestkey]", "Test.Test1[badtestkey]", "test1[badtestkey]", "Test1[badtestkey]", "eq", "eq")
	AssertDeepError(t, errs, "Test.test1[badtestkey]", "Test.Test1[badtestkey]", "test1[badtestkey]", "Test1[badtestkey]", "eq", "eq")
	AssertDeepError(t, errs, "Test.test2[10]", "Test.Test2[10]", "test2[10]", "Test2[10]", "eq", "eq")
	AssertDeepError(t, errs, "Test.test2[10]", "Test.Test2[10]", "test2[10]", "Test2[10]", "eq", "eq")
}

// Thanks @adrian-sgn specific test for your specific scenario
func TestKeysCustomValidation(t *testing.T) {
	type LangCode string
	type Label map[LangCode]string
	type TestMapStructPtr struct {
		Label Label `validate:"dive,keys,lang_code,endkeys,required"`
	}

	validate := New()
	err := validate.RegisterValidation("lang_code", func(fl FieldLevel) bool {
		validLangCodes := map[LangCode]struct{}{
			"en": {},
			"es": {},
			"pt": {},
		}

		_, ok := validLangCodes[fl.Field().Interface().(LangCode)]
		return ok
	})
	Equal(t, err, nil)

	label := Label{
		"en":  "Good morning!",
		"pt":  "",
		"es":  "¡Buenos días!",
		"xx":  "Bad key",
		"xxx": "",
	}

	err = validate.Struct(TestMapStructPtr{label})
	NotEqual(t, err, nil)

	errs := err.(ValidationErrors)
	Equal(t, len(errs), 4)

	AssertDeepError(t, errs, "TestMapStructPtr.Label[xx]", "TestMapStructPtr.Label[xx]", "Label[xx]", "Label[xx]", "lang_code", "lang_code")
	AssertDeepError(t, errs, "TestMapStructPtr.Label[pt]", "TestMapStructPtr.Label[pt]", "Label[pt]", "Label[pt]", "required", "required")
	AssertDeepError(t, errs, "TestMapStructPtr.Label[xxx]", "TestMapStructPtr.Label[xxx]", "Label[xxx]", "Label[xxx]", "lang_code", "lang_code")
	AssertDeepError(t, errs, "TestMapStructPtr.Label[xxx]", "TestMapStructPtr.Label[xxx]", "Label[xxx]", "Label[xxx]", "required", "required")

	// find specific error
	var e FieldError
	for _, e = range errs {
		if e.Namespace() == "TestMapStructPtr.Label[xxx]" {
			break
		}
	}

	Equal(t, e.Param(), "")
	Equal(t, e.Value().(LangCode), LangCode("xxx"))

	for _, e = range errs {
		if e.Namespace() == "TestMapStructPtr.Label[xxx]" && e.Tag() == "required" {
			break
		}
	}

	Equal(t, e.Param(), "")
	Equal(t, e.Value().(string), "")
}

func TestKeyOrs(t *testing.T) {
	type Test struct {
		Test1 map[string]string `validate:"gt=0,dive,keys,eq=testkey|eq=testkeyok,endkeys,eq=testval" json:"test1"`
	}

	var tst Test
	validate := New()
	err := validate.Struct(tst)
	NotEqual(t, err, nil)
	Equal(t, len(err.(ValidationErrors)), 1)
	AssertError(t, err.(ValidationErrors), "Test.Test1", "Test.Test1", "Test1", "Test1", "gt")

	tst.Test1 = map[string]string{
		"testkey": "testval",
	}

	err = validate.Struct(tst)
	Equal(t, err, nil)

	tst.Test1["badtestkey"] = "badtestval"
	err = validate.Struct(tst)
	NotEqual(t, err, nil)

	errs := err.(ValidationErrors)
	Equal(t, len(errs), 2)

	AssertDeepError(t, errs, "Test.Test1[badtestkey]", "Test.Test1[badtestkey]", "Test1[badtestkey]", "Test1[badtestkey]", "eq=testkey|eq=testkeyok", "eq=testkey|eq=testkeyok")
	AssertDeepError(t, errs, "Test.Test1[badtestkey]", "Test.Test1[badtestkey]", "Test1[badtestkey]", "Test1[badtestkey]", "eq", "eq")

	validate.RegisterAlias("okkey", "eq=testkey|eq=testkeyok")
	type Test2 struct {
		Test1 map[string]string `validate:"gt=0,dive,keys,okkey,endkeys,eq=testval" json:"test1"`
	}

	var tst2 Test2
	err = validate.Struct(tst2)
	NotEqual(t, err, nil)
	Equal(t, len(err.(ValidationErrors)), 1)
	AssertError(t, err.(ValidationErrors), "Test2.Test1", "Test2.Test1", "Test1", "Test1", "gt")

	tst2.Test1 = map[string]string{
		"testkey": "testval",
	}

	err = validate.Struct(tst2)
	Equal(t, err, nil)

	tst2.Test1["badtestkey"] = "badtestval"

	err = validate.Struct(tst2)
	NotEqual(t, err, nil)

	errs = err.(ValidationErrors)
	Equal(t, len(errs), 2)

	AssertDeepError(t, errs, "Test2.Test1[badtestkey]", "Test2.Test1[badtestkey]", "Test1[badtestkey]", "Test1[badtestkey]", "okkey", "eq=testkey|eq=testkeyok")
	AssertDeepError(t, errs, "Test2.Test1[badtestkey]", "Test2.Test1[badtestkey]", "Test1[badtestkey]", "Test1[badtestkey]", "eq", "eq")
}

func TestRequiredIf(t *testing.T) {
	type Inner struct {
		Field *string
	}

	fieldVal := "test"
	test := struct {
		Inner   *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"required_if=FieldE test" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"required_if=Field1 test" json:"field_2"`
		Field3  map[string]string `validate:"required_if=Field2 test" json:"field_3"`
		Field4  interface{}       `validate:"required_if=Field3 1" json:"field_4"`
		Field5  int               `validate:"required_if=Inner.Field test" json:"field_5"`
		Field6  uint              `validate:"required_if=Field5 1" json:"field_6"`
		Field7  float32           `validate:"required_if=Field6 1" json:"field_7"`
		Field8  float64           `validate:"required_if=Field7 1.0" json:"field_8"`
		Field9  Inner             `validate:"required_if=Field1 test" json:"field_9"`
		Field10 *Inner            `validate:"required_if=Field1 test" json:"field_10"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: 2,
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner   *Inner
		Inner2  *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"required_if=FieldE test" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"required_if=Field1 test" json:"field_2"`
		Field3  map[string]string `validate:"required_if=Field2 test" json:"field_3"`
		Field4  interface{}       `validate:"required_if=Field2 test" json:"field_4"`
		Field5  string            `validate:"required_if=Field3 1" json:"field_5"`
		Field6  string            `validate:"required_if=Inner.Field test" json:"field_6"`
		Field7  string            `validate:"required_if=Inner2.Field test" json:"field_7"`
		Field8  Inner             `validate:"required_if=Field2 test" json:"field_8"`
		Field9  *Inner            `validate:"required_if=Field2 test" json:"field_9"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field2: &fieldVal,
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 5)
	AssertError(t, errs, "Field3", "Field3", "Field3", "Field3", "required_if")
	AssertError(t, errs, "Field4", "Field4", "Field4", "Field4", "required_if")
	AssertError(t, errs, "Field6", "Field6", "Field6", "Field6", "required_if")
	AssertError(t, errs, "Field8", "Field8", "Field8", "Field8", "required_if")
	AssertError(t, errs, "Field9", "Field9", "Field9", "Field9", "required_if")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("test3 should have panicked!")
		}
	}()

	test3 := struct {
		Inner  *Inner
		Field1 string `validate:"required_if=Inner.Field" json:"field_1"`
	}{
		Inner: &Inner{Field: &fieldVal},
	}
	_ = validate.Struct(test3)
}

func TestRequiredUnless(t *testing.T) {
	type Inner struct {
		Field *string
	}

	fieldVal := "test"
	test := struct {
		Inner   *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"required_unless=FieldE test" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"required_unless=Field1 test" json:"field_2"`
		Field3  map[string]string `validate:"required_unless=Field2 test" json:"field_3"`
		Field4  interface{}       `validate:"required_unless=Field3 1" json:"field_4"`
		Field5  int               `validate:"required_unless=Inner.Field test" json:"field_5"`
		Field6  uint              `validate:"required_unless=Field5 2" json:"field_6"`
		Field7  float32           `validate:"required_unless=Field6 0" json:"field_7"`
		Field8  float64           `validate:"required_unless=Field7 0.0" json:"field_8"`
		Field9  bool              `validate:"omitempty" json:"field_9"`
		Field10 string            `validate:"required_unless=Field9 true" json:"field_10"`
		Field11 Inner             `validate:"required_unless=Field9 true" json:"field_11"`
		Field12 *Inner            `validate:"required_unless=Field9 true" json:"field_12"`
	}{
		FieldE: "test",
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: 2,
		Field9: true,
	}
	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner   *Inner
		Inner2  *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"required_unless=FieldE test" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"required_unless=Field1 test" json:"field_2"`
		Field3  map[string]string `validate:"required_unless=Field2 test" json:"field_3"`
		Field4  interface{}       `validate:"required_unless=Field2 test" json:"field_4"`
		Field5  string            `validate:"required_unless=Field3 0" json:"field_5"`
		Field6  string            `validate:"required_unless=Inner.Field test" json:"field_6"`
		Field7  string            `validate:"required_unless=Inner2.Field test" json:"field_7"`
		Field8  bool              `validate:"omitempty" json:"field_8"`
		Field9  string            `validate:"required_unless=Field8 true" json:"field_9"`
		Field10 Inner             `validate:"required_unless=Field9 true" json:"field_10"`
		Field11 *Inner            `validate:"required_unless=Field9 true" json:"field_11"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		FieldE: "test",
		Field1: "test",
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 6)
	AssertError(t, errs, "Field3", "Field3", "Field3", "Field3", "required_unless")
	AssertError(t, errs, "Field4", "Field4", "Field4", "Field4", "required_unless")
	AssertError(t, errs, "Field7", "Field7", "Field7", "Field7", "required_unless")
	AssertError(t, errs, "Field9", "Field9", "Field9", "Field9", "required_unless")
	AssertError(t, errs, "Field10", "Field10", "Field10", "Field10", "required_unless")
	AssertError(t, errs, "Field11", "Field11", "Field11", "Field11", "required_unless")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("test3 should have panicked!")
		}
	}()

	test3 := struct {
		Inner  *Inner
		Field1 string `validate:"required_unless=Inner.Field" json:"field_1"`
	}{
		Inner: &Inner{Field: &fieldVal},
	}
	_ = validate.Struct(test3)
}

func TestSkipUnless(t *testing.T) {
	type Inner struct {
		Field *string
	}

	fieldVal := "test1"
	test := struct {
		Inner   *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"skip_unless=FieldE test" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"skip_unless=Field1 test" json:"field_2"`
		Field3  map[string]string `validate:"skip_unless=Field2 test" json:"field_3"`
		Field4  interface{}       `validate:"skip_unless=Field3 1" json:"field_4"`
		Field5  int               `validate:"skip_unless=Inner.Field test" json:"field_5"`
		Field6  uint              `validate:"skip_unless=Field5 2" json:"field_6"`
		Field7  float32           `validate:"skip_unless=Field6 1" json:"field_7"`
		Field8  float64           `validate:"skip_unless=Field7 1.0" json:"field_8"`
		Field9  bool              `validate:"omitempty" json:"field_9"`
		Field10 string            `validate:"skip_unless=Field9 false" json:"field_10"`
		Field11 Inner             `validate:"skip_unless=Field9 false" json:"field_11"`
		Field12 *Inner            `validate:"skip_unless=Field9 false" json:"field_12"`
	}{
		FieldE: "test1",
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: 3,
		Field9: true,
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner   *Inner
		Inner2  *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"skip_unless=FieldE test" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"skip_unless=Field1 test" json:"field_2"`
		Field3  map[string]string `validate:"skip_unless=Field2 test" json:"field_3"`
		Field4  interface{}       `validate:"skip_unless=Field2 test" json:"field_4"`
		Field5  string            `validate:"skip_unless=Field3 0" json:"field_5"`
		Field6  string            `validate:"skip_unless=Inner.Field test" json:"field_6"`
		Field7  string            `validate:"skip_unless=Inner2.Field test" json:"field_7"`
		Field8  bool              `validate:"omitempty" json:"field_8"`
		Field9  string            `validate:"skip_unless=Field8 true" json:"field_9"`
		Field10 Inner             `validate:"skip_unless=Field8 false" json:"field_10"`
		Field11 *Inner            `validate:"skip_unless=Field8 false" json:"field_11"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		FieldE: "test1",
		Field1: "test1",
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 3)
	AssertError(t, errs, "Field5", "Field5", "Field5", "Field5", "skip_unless")
	AssertError(t, errs, "Field10", "Field10", "Field10", "Field10", "skip_unless")
	AssertError(t, errs, "Field11", "Field11", "Field11", "Field11", "skip_unless")

	test3 := struct {
		Inner  *Inner
		Field1 string `validate:"skip_unless=Inner.Field" json:"field_1"`
	}{
		Inner: &Inner{Field: &fieldVal},
	}
	PanicMatches(t, func() {
		_ = validate.Struct(test3)
	}, "Bad param number for skip_unless Field1")

	test4 := struct {
		Inner  *Inner
		Field1 string `validate:"skip_unless=Inner.Field test1" json:"field_1"`
	}{
		Inner: &Inner{Field: &fieldVal},
	}
	errs = validate.Struct(test4)
	NotEqual(t, errs, nil)

	ve = errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "Field1", "Field1", "Field1", "Field1", "skip_unless")
}

func TestRequiredWith(t *testing.T) {
	type Inner struct {
		Field *string
	}

	fieldVal := "test"
	test := struct {
		Inner   *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"required_with=FieldE" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"required_with=Field1" json:"field_2"`
		Field3  map[string]string `validate:"required_with=Field2" json:"field_3"`
		Field4  interface{}       `validate:"required_with=Field3" json:"field_4"`
		Field5  string            `validate:"required_with=Field" json:"field_5"`
		Field6  Inner             `validate:"required_with=Field2" json:"field_6"`
		Field7  *Inner            `validate:"required_with=Field2" json:"field_7"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: Inner{Field: &fieldVal},
		Field7: &Inner{Field: &fieldVal},
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner   *Inner
		Inner2  *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"required_with=FieldE" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"required_with=Field1" json:"field_2"`
		Field3  map[string]string `validate:"required_with=Field2" json:"field_3"`
		Field4  interface{}       `validate:"required_with=Field2" json:"field_4"`
		Field5  string            `validate:"required_with=Field3" json:"field_5"`
		Field6  string            `validate:"required_with=Inner.Field" json:"field_6"`
		Field7  string            `validate:"required_with=Inner2.Field" json:"field_7"`
		Field8  Inner             `validate:"required_with=Field2" json:"field_8"`
		Field9  *Inner            `validate:"required_with=Field2" json:"field_9"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field2: &fieldVal,
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 5)
	AssertError(t, errs, "Field3", "Field3", "Field3", "Field3", "required_with")
	AssertError(t, errs, "Field4", "Field4", "Field4", "Field4", "required_with")
	AssertError(t, errs, "Field6", "Field6", "Field6", "Field6", "required_with")
	AssertError(t, errs, "Field8", "Field8", "Field8", "Field8", "required_with")
	AssertError(t, errs, "Field9", "Field9", "Field9", "Field9", "required_with")
}

func TestExcludedWith(t *testing.T) {
	type Inner struct {
		FieldE string
		Field  *string
	}

	fieldVal := "test"
	test := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_with=FieldE" json:"field_1"`
		Field2 *string           `validate:"excluded_with=FieldE" json:"field_2"`
		Field3 map[string]string `validate:"excluded_with=FieldE" json:"field_3"`
		Field4 interface{}       `validate:"excluded_with=FieldE" json:"field_4"`
		Field5 string            `validate:"excluded_with=Inner.FieldE" json:"field_5"`
		Field6 string            `validate:"excluded_with=Inner2.FieldE" json:"field_6"`
		Field7 Inner             `validate:"excluded_with=FieldE" json:"field_7"`
		Field8 *Inner            `validate:"excluded_with=FieldE" json:"field_8"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field1: fieldVal,
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: "test",
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_with=Field" json:"field_1"`
		Field2 *string           `validate:"excluded_with=Field" json:"field_2"`
		Field3 map[string]string `validate:"excluded_with=Field" json:"field_3"`
		Field4 interface{}       `validate:"excluded_with=Field" json:"field_4"`
		Field5 string            `validate:"excluded_with=Inner.Field" json:"field_5"`
		Field6 string            `validate:"excluded_with=Inner2.Field" json:"field_6"`
		Field7 Inner             `validate:"excluded_with=Field" json:"field_7"`
		Field8 *Inner            `validate:"excluded_with=Field" json:"field_8"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field:  "populated",
		Field1: fieldVal,
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: "test",
		Field7: Inner{FieldE: "potato"},
		Field8: &Inner{FieldE: "potato"},
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 7)
	for i := 1; i <= 7; i++ {
		// accounting for field 7 & 8 failures, 6 skipped because no failure
		if i > 5 {
			i++
		}
		name := fmt.Sprintf("Field%d", i)
		AssertError(t, errs, name, name, name, name, "excluded_with")
	}

	test3 := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_with=FieldE" json:"field_1"`
		Field2 *string           `validate:"excluded_with=FieldE" json:"field_2"`
		Field3 map[string]string `validate:"excluded_with=FieldE" json:"field_3"`
		Field4 interface{}       `validate:"excluded_with=FieldE" json:"field_4"`
		Field5 string            `validate:"excluded_with=Inner.FieldE" json:"field_5"`
		Field6 string            `validate:"excluded_with=Inner2.FieldE" json:"field_6"`
		Field7 Inner             `validate:"excluded_with=FieldE" json:"field_7"`
		Field8 *Inner            `validate:"excluded_with=FieldE" json:"field_8"`
	}{
		Inner:  &Inner{FieldE: "populated"},
		Inner2: &Inner{FieldE: "populated"},
		FieldE: "populated",
	}

	validate = New()
	errs = validate.Struct(test3)
	Equal(t, errs, nil)
}

func TestExcludedWithout(t *testing.T) {
	type Inner struct {
		FieldE string
		Field  *string
	}

	fieldVal := "test"
	test := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_without=Field" json:"field_1"`
		Field2 *string           `validate:"excluded_without=Field" json:"field_2"`
		Field3 map[string]string `validate:"excluded_without=Field" json:"field_3"`
		Field4 interface{}       `validate:"excluded_without=Field" json:"field_4"`
		Field5 string            `validate:"excluded_without=Inner.Field" json:"field_5"`
		Field6 Inner             `validate:"excluded_without=Field" json:"field_6"`
		Field7 *Inner            `validate:"excluded_without=Field" json:"field_7"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field:  "populated",
		Field1: fieldVal,
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: Inner{FieldE: "potato"},
		Field7: &Inner{FieldE: "potato"},
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_without=FieldE" json:"field_1"`
		Field2 *string           `validate:"excluded_without=FieldE" json:"field_2"`
		Field3 map[string]string `validate:"excluded_without=FieldE" json:"field_3"`
		Field4 interface{}       `validate:"excluded_without=FieldE" json:"field_4"`
		Field5 string            `validate:"excluded_without=Inner.FieldE" json:"field_5"`
		Field6 string            `validate:"excluded_without=Inner2.FieldE" json:"field_6"`
		Field7 Inner             `validate:"excluded_without=FieldE" json:"field_7"`
		Field8 *Inner            `validate:"excluded_without=FieldE" json:"field_8"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field1: fieldVal,
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: "test",
		Field7: Inner{FieldE: "potato"},
		Field8: &Inner{FieldE: "potato"},
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 8)
	for i := 1; i <= 8; i++ {
		name := fmt.Sprintf("Field%d", i)
		AssertError(t, errs, name, name, name, name, "excluded_without")
	}

	test3 := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_without=Field" json:"field_1"`
		Field2 *string           `validate:"excluded_without=Field" json:"field_2"`
		Field3 map[string]string `validate:"excluded_without=Field" json:"field_3"`
		Field4 interface{}       `validate:"excluded_without=Field" json:"field_4"`
		Field5 string            `validate:"excluded_without=Inner.Field" json:"field_5"`
		Field6 Inner             `validate:"excluded_without=Field" json:"field_6"`
		Field7 *Inner            `validate:"excluded_without=Field" json:"field_7"`
	}{
		Inner: &Inner{Field: &fieldVal},
		Field: "populated",
	}

	validate = New()

	errs = validate.Struct(test3)
	Equal(t, errs, nil)
}

func TestExcludedWithAll(t *testing.T) {
	type Inner struct {
		FieldE string
		Field  *string
	}

	fieldVal := "test"
	test := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_with_all=FieldE Field" json:"field_1"`
		Field2 *string           `validate:"excluded_with_all=FieldE Field" json:"field_2"`
		Field3 map[string]string `validate:"excluded_with_all=FieldE Field" json:"field_3"`
		Field4 interface{}       `validate:"excluded_with_all=FieldE Field" json:"field_4"`
		Field5 string            `validate:"excluded_with_all=Inner.FieldE" json:"field_5"`
		Field6 string            `validate:"excluded_with_all=Inner2.FieldE" json:"field_6"`
		Field7 Inner             `validate:"excluded_with_all=FieldE Field" json:"field_7"`
		Field8 *Inner            `validate:"excluded_with_all=FieldE Field" json:"field_8"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field:  fieldVal,
		Field1: fieldVal,
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: "test",
		Field7: Inner{FieldE: "potato"},
		Field8: &Inner{FieldE: "potato"},
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_with_all=Field FieldE" json:"field_1"`
		Field2 *string           `validate:"excluded_with_all=Field FieldE" json:"field_2"`
		Field3 map[string]string `validate:"excluded_with_all=Field FieldE" json:"field_3"`
		Field4 interface{}       `validate:"excluded_with_all=Field FieldE" json:"field_4"`
		Field5 string            `validate:"excluded_with_all=Inner.Field" json:"field_5"`
		Field6 string            `validate:"excluded_with_all=Inner2.Field" json:"field_6"`
		Field7 Inner             `validate:"excluded_with_all=Field FieldE" json:"field_7"`
		Field8 *Inner            `validate:"excluded_with_all=Field FieldE" json:"field_8"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field:  "populated",
		FieldE: "populated",
		Field1: fieldVal,
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: "test",
		Field7: Inner{FieldE: "potato"},
		Field8: &Inner{FieldE: "potato"},
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 7)
	for i := 1; i <= 7; i++ {
		// accounting for no err for field 6
		if i > 5 {
			i++
		}
		name := fmt.Sprintf("Field%d", i)
		AssertError(t, errs, name, name, name, name, "excluded_with_all")
	}

	test3 := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_with_all=FieldE Field" json:"field_1"`
		Field2 *string           `validate:"excluded_with_all=FieldE Field" json:"field_2"`
		Field3 map[string]string `validate:"excluded_with_all=FieldE Field" json:"field_3"`
		Field4 interface{}       `validate:"excluded_with_all=FieldE Field" json:"field_4"`
		Field5 string            `validate:"excluded_with_all=Inner.FieldE" json:"field_5"`
		Field6 string            `validate:"excluded_with_all=Inner2.FieldE" json:"field_6"`
		Field7 Inner             `validate:"excluded_with_all=Field FieldE" json:"field_7"`
		Field8 *Inner            `validate:"excluded_with_all=Field FieldE" json:"field_8"`
	}{
		Inner:  &Inner{FieldE: "populated"},
		Inner2: &Inner{FieldE: "populated"},
		Field:  "populated",
		FieldE: "populated",
	}

	validate = New()
	errs = validate.Struct(test3)
	Equal(t, errs, nil)
}

func TestExcludedWithoutAll(t *testing.T) {
	type Inner struct {
		FieldE string
		Field  *string
	}

	fieldVal := "test"
	test := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_without_all=Field FieldE" json:"field_1"`
		Field2 *string           `validate:"excluded_without_all=Field FieldE" json:"field_2"`
		Field3 map[string]string `validate:"excluded_without_all=Field FieldE" json:"field_3"`
		Field4 interface{}       `validate:"excluded_without_all=Field FieldE" json:"field_4"`
		Field5 string            `validate:"excluded_without_all=Inner.Field Inner2.Field" json:"field_5"`
		Field6 Inner             `validate:"excluded_without_all=Field FieldE" json:"field_6"`
		Field7 *Inner            `validate:"excluded_without_all=Field FieldE" json:"field_7"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Inner2: &Inner{Field: &fieldVal},
		Field:  "populated",
		Field1: fieldVal,
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: Inner{FieldE: "potato"},
		Field7: &Inner{FieldE: "potato"},
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_without_all=FieldE Field" json:"field_1"`
		Field2 *string           `validate:"excluded_without_all=FieldE Field" json:"field_2"`
		Field3 map[string]string `validate:"excluded_without_all=FieldE Field" json:"field_3"`
		Field4 interface{}       `validate:"excluded_without_all=FieldE Field" json:"field_4"`
		Field5 string            `validate:"excluded_without_all=Inner.FieldE" json:"field_5"`
		Field6 string            `validate:"excluded_without_all=Inner2.FieldE" json:"field_6"`
		Field7 Inner             `validate:"excluded_without_all=Field FieldE" json:"field_7"`
		Field8 *Inner            `validate:"excluded_without_all=Field FieldE" json:"field_8"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field1: fieldVal,
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: "test",
		Field7: Inner{FieldE: "potato"},
		Field8: &Inner{FieldE: "potato"},
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 8)
	for i := 1; i <= 8; i++ {
		name := fmt.Sprintf("Field%d", i)
		AssertError(t, errs, name, name, name, name, "excluded_without_all")
	}

	test3 := struct {
		Inner  *Inner
		Inner2 *Inner
		Field  string            `validate:"omitempty" json:"field"`
		FieldE string            `validate:"omitempty" json:"field_e"`
		Field1 string            `validate:"excluded_without_all=Field FieldE" json:"field_1"`
		Field2 *string           `validate:"excluded_without_all=Field FieldE" json:"field_2"`
		Field3 map[string]string `validate:"excluded_without_all=Field FieldE" json:"field_3"`
		Field4 interface{}       `validate:"excluded_without_all=Field FieldE" json:"field_4"`
		Field5 string            `validate:"excluded_without_all=Inner.Field Inner2.Field" json:"field_5"`
		Field6 Inner             `validate:"excluded_without_all=Field FieldE" json:"field_6"`
		Field7 *Inner            `validate:"excluded_without_all=Field FieldE" json:"field_7"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Inner2: &Inner{Field: &fieldVal},
		Field:  "populated",
		FieldE: "populated",
	}

	validate = New()

	errs = validate.Struct(test3)
	Equal(t, errs, nil)
}

func TestRequiredWithAll(t *testing.T) {
	type Inner struct {
		Field *string
	}

	fieldVal := "test"
	test := struct {
		Inner   *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"required_with_all=FieldE" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"required_with_all=Field1" json:"field_2"`
		Field3  map[string]string `validate:"required_with_all=Field2" json:"field_3"`
		Field4  interface{}       `validate:"required_with_all=Field3" json:"field_4"`
		Field5  string            `validate:"required_with_all=Inner.Field" json:"field_5"`
		Field6  Inner             `validate:"required_with_all=Field1 Field2" json:"field_6"`
		Field7  *Inner            `validate:"required_with_all=Field1 Field2" json:"field_7"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field1: "test_field1",
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: Inner{Field: &fieldVal},
		Field7: &Inner{Field: &fieldVal},
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner   *Inner
		Inner2  *Inner
		FieldE  string            `validate:"omitempty" json:"field_e"`
		FieldER string            `validate:"required_with_all=FieldE" json:"field_er"`
		Field1  string            `validate:"omitempty" json:"field_1"`
		Field2  *string           `validate:"required_with_all=Field1" json:"field_2"`
		Field3  map[string]string `validate:"required_with_all=Field2" json:"field_3"`
		Field4  interface{}       `validate:"required_with_all=Field1 FieldE" json:"field_4"`
		Field5  string            `validate:"required_with_all=Inner.Field Field2" json:"field_5"`
		Field6  string            `validate:"required_with_all=Inner2.Field Field2" json:"field_6"`
		Field7  Inner             `validate:"required_with_all=Inner.Field Field2" json:"field_7"`
		Field8  *Inner            `validate:"required_with_all=Inner.Field Field2" json:"field_8"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field2: &fieldVal,
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 4)
	AssertError(t, errs, "Field3", "Field3", "Field3", "Field3", "required_with_all")
	AssertError(t, errs, "Field5", "Field5", "Field5", "Field5", "required_with_all")
	AssertError(t, errs, "Field7", "Field7", "Field7", "Field7", "required_with_all")
	AssertError(t, errs, "Field8", "Field8", "Field8", "Field8", "required_with_all")
}

func TestRequiredWithout(t *testing.T) {
	type Inner struct {
		Field *string
	}

	fieldVal := "test"
	test := struct {
		Inner  *Inner
		Field1 string            `validate:"omitempty" json:"field_1"`
		Field2 *string           `validate:"required_without=Field1" json:"field_2"`
		Field3 map[string]string `validate:"required_without=Field2" json:"field_3"`
		Field4 interface{}       `validate:"required_without=Field3" json:"field_4"`
		Field5 string            `validate:"required_without=Field3" json:"field_5"`
		Field6 Inner             `validate:"required_without=Field1" json:"field_6"`
		Field7 *Inner            `validate:"required_without=Field1" json:"field_7"`
	}{
		Inner:  &Inner{Field: &fieldVal},
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: Inner{Field: &fieldVal},
		Field7: &Inner{Field: &fieldVal},
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Inner   *Inner
		Inner2  *Inner
		Field1  string            `json:"field_1"`
		Field2  *string           `validate:"required_without=Field1" json:"field_2"`
		Field3  map[string]string `validate:"required_without=Field2" json:"field_3"`
		Field4  interface{}       `validate:"required_without=Field3" json:"field_4"`
		Field5  string            `validate:"required_without=Field3" json:"field_5"`
		Field6  string            `validate:"required_without=Field1" json:"field_6"`
		Field7  string            `validate:"required_without=Inner.Field" json:"field_7"`
		Field8  string            `validate:"required_without=Inner.Field" json:"field_8"`
		Field9  Inner             `validate:"required_without=Field1" json:"field_9"`
		Field10 *Inner            `validate:"required_without=Field1" json:"field_10"`
	}{
		Inner:  &Inner{},
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
	}

	errs = validate.Struct(&test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 6)
	AssertError(t, errs, "Field2", "Field2", "Field2", "Field2", "required_without")
	AssertError(t, errs, "Field6", "Field6", "Field6", "Field6", "required_without")
	AssertError(t, errs, "Field7", "Field7", "Field7", "Field7", "required_without")
	AssertError(t, errs, "Field8", "Field8", "Field8", "Field8", "required_without")
	AssertError(t, errs, "Field9", "Field9", "Field9", "Field9", "required_without")
	AssertError(t, errs, "Field10", "Field10", "Field10", "Field10", "required_without")

	test3 := struct {
		Field1 *string `validate:"required_without=Field2,omitempty,min=1" json:"field_1"`
		Field2 *string `validate:"required_without=Field1,omitempty,min=1" json:"field_2"`
	}{
		Field1: &fieldVal,
	}

	errs = validate.Struct(&test3)
	Equal(t, errs, nil)

	test4 := struct {
		Field1 string `validate:"required_without=Field2 Field3,omitempty,min=1" json:"field_1"`
		Field2 string `json:"field_2"`
		Field3 string `json:"field_3"`
	}{
		Field1: "test",
	}

	errs = validate.Struct(&test4)
	Equal(t, errs, nil)

	test5 := struct {
		Field1 string `validate:"required_without=Field2 Field3,omitempty,min=1" json:"field_1"`
		Field2 string `json:"field_2"`
		Field3 string `json:"field_3"`
	}{
		Field3: "test",
	}

	errs = validate.Struct(&test5)
	NotEqual(t, errs, nil)

	ve = errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "Field1", "Field1", "Field1", "Field1", "required_without")
}

func TestRequiredWithoutAll(t *testing.T) {
	type nested struct {
		value string
	}
	fieldVal := "test"
	test := struct {
		Field1 string            `validate:"omitempty" json:"field_1"`
		Field2 *string           `validate:"required_without_all=Field1" json:"field_2"`
		Field3 map[string]string `validate:"required_without_all=Field2" json:"field_3"`
		Field4 interface{}       `validate:"required_without_all=Field3" json:"field_4"`
		Field5 string            `validate:"required_without_all=Field3" json:"field_5"`
		Field6 nested            `validate:"required_without_all=Field1" json:"field_6"`
		Field7 *nested           `validate:"required_without_all=Field1" json:"field_7"`
	}{
		Field1: "",
		Field2: &fieldVal,
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
		Field6: nested{"potato"},
		Field7: &nested{"potato"},
	}

	validate := New()
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		Field1 string            `validate:"omitempty" json:"field_1"`
		Field2 *string           `validate:"required_without_all=Field1" json:"field_2"`
		Field3 map[string]string `validate:"required_without_all=Field2" json:"field_3"`
		Field4 interface{}       `validate:"required_without_all=Field3" json:"field_4"`
		Field5 string            `validate:"required_without_all=Field3" json:"field_5"`
		Field6 string            `validate:"required_without_all=Field1 Field3" json:"field_6"`
		Field7 nested            `validate:"required_without_all=Field1" json:"field_7"`
		Field8 *nested           `validate:"required_without_all=Field1" json:"field_8"`
	}{
		Field3: map[string]string{"key": "val"},
		Field4: "test",
		Field5: "test",
	}

	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)

	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 3)
	AssertError(t, errs, "Field2", "Field2", "Field2", "Field2", "required_without_all")
	AssertError(t, errs, "Field7", "Field7", "Field7", "Field7", "required_without_all")
	AssertError(t, errs, "Field8", "Field8", "Field8", "Field8", "required_without_all")
}

func TestExcludedIf(t *testing.T) {
	validate := New()
	type Inner struct {
		Field *string
	}

	shouldExclude := "exclude"
	shouldNotExclude := "dontExclude"
	test1 := struct {
		FieldE  string  `validate:"omitempty" json:"field_e"`
		FieldER *string `validate:"excluded_if=FieldE exclude" json:"field_er"`
	}{
		FieldE: shouldExclude,
	}
	errs := validate.Struct(test1)
	Equal(t, errs, nil)

	test2 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_if=FieldE exclude" json:"field_er"`
	}{
		FieldE:  shouldExclude,
		FieldER: "set",
	}
	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)
	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "FieldER", "FieldER", "FieldER", "FieldER", "excluded_if")

	test3 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldF  string `validate:"omitempty" json:"field_f"`
		FieldER string `validate:"excluded_if=FieldE exclude FieldF exclude" json:"field_er"`
	}{
		FieldE:  shouldExclude,
		FieldF:  shouldExclude,
		FieldER: "set",
	}
	errs = validate.Struct(test3)
	NotEqual(t, errs, nil)
	ve = errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "FieldER", "FieldER", "FieldER", "FieldER", "excluded_if")

	test4 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldF  string `validate:"omitempty" json:"field_f"`
		FieldER string `validate:"excluded_if=FieldE exclude FieldF exclude" json:"field_er"`
	}{
		FieldE:  shouldExclude,
		FieldF:  shouldNotExclude,
		FieldER: "set",
	}
	errs = validate.Struct(test4)
	Equal(t, errs, nil)

	test5 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_if=FieldE exclude" json:"field_er"`
	}{
		FieldE: shouldNotExclude,
	}
	errs = validate.Struct(test5)
	Equal(t, errs, nil)

	test6 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_if=FieldE exclude" json:"field_er"`
	}{
		FieldE:  shouldNotExclude,
		FieldER: "set",
	}
	errs = validate.Struct(test6)
	Equal(t, errs, nil)

	test7 := struct {
		Inner  *Inner
		FieldE string `validate:"omitempty" json:"field_e"`
		Field1 int    `validate:"excluded_if=Inner.Field exclude" json:"field_1"`
	}{
		Inner: &Inner{Field: &shouldExclude},
	}
	errs = validate.Struct(test7)
	Equal(t, errs, nil)

	test8 := struct {
		Inner  *Inner
		FieldE string `validate:"omitempty" json:"field_e"`
		Field1 int    `validate:"excluded_if=Inner.Field exclude" json:"field_1"`
	}{
		Inner:  &Inner{Field: &shouldExclude},
		Field1: 1,
	}
	errs = validate.Struct(test8)
	NotEqual(t, errs, nil)
	ve = errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "Field1", "Field1", "Field1", "Field1", "excluded_if")

	test9 := struct {
		Inner  *Inner
		FieldE string `validate:"omitempty" json:"field_e"`
		Field1 int    `validate:"excluded_if=Inner.Field exclude" json:"field_1"`
	}{
		Inner: &Inner{Field: &shouldNotExclude},
	}
	errs = validate.Struct(test9)
	Equal(t, errs, nil)

	test10 := struct {
		Inner  *Inner
		FieldE string `validate:"omitempty" json:"field_e"`
		Field1 int    `validate:"excluded_if=Inner.Field exclude" json:"field_1"`
	}{
		Inner:  &Inner{Field: &shouldNotExclude},
		Field1: 1,
	}
	errs = validate.Struct(test10)
	Equal(t, errs, nil)

	test11 := struct {
		Field1 bool
		Field2 *string `validate:"excluded_if=Field1 false"`
	}{
		Field1: false,
		Field2: nil,
	}
	errs = validate.Struct(test11)
	Equal(t, errs, nil)

	test12 := struct {
		Field1 bool
		Field2 *string `validate:"excluded_if=Field1 !Field1"`
	}{
		Field1: true,
		Field2: nil,
	}
	errs = validate.Struct(test12)
	Equal(t, errs, nil)
	// Checks number of params in struct tag is correct
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("panicTest should have panicked!")
		}
	}()
	fieldVal := "panicTest"
	panicTest := struct {
		Inner  *Inner
		Field1 string `validate:"excluded_if=Inner.Field" json:"field_1"`
	}{
		Inner: &Inner{Field: &fieldVal},
	}
	_ = validate.Struct(panicTest)
}

func TestExcludedUnless(t *testing.T) {
	validate := New()
	type Inner struct {
		Field *string
	}

	fieldVal := "test"
	test := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_unless=FieldE test" json:"field_er"`
	}{
		FieldE:  "test",
		FieldER: "filled",
	}
	errs := validate.Struct(test)
	Equal(t, errs, nil)

	test2 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_unless=FieldE test" json:"field_er"`
	}{
		FieldE:  "notest",
		FieldER: "filled",
	}
	errs = validate.Struct(test2)
	NotEqual(t, errs, nil)
	ve := errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "FieldER", "FieldER", "FieldER", "FieldER", "excluded_unless")

	// test5 and test6: excluded_unless has no effect if FieldER is left blank
	test5 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_unless=FieldE test" json:"field_er"`
	}{
		FieldE: "test",
	}
	errs = validate.Struct(test5)
	Equal(t, errs, nil)

	test6 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_unless=FieldE test" json:"field_er"`
	}{
		FieldE: "notest",
	}
	errs = validate.Struct(test6)
	Equal(t, errs, nil)

	shouldError := "notest"
	test3 := struct {
		Inner  *Inner
		Field1 string `validate:"excluded_unless=Inner.Field test" json:"field_1"`
	}{
		Inner:  &Inner{Field: &shouldError},
		Field1: "filled",
	}
	errs = validate.Struct(test3)
	NotEqual(t, errs, nil)
	ve = errs.(ValidationErrors)
	Equal(t, len(ve), 1)
	AssertError(t, errs, "Field1", "Field1", "Field1", "Field1", "excluded_unless")

	shouldPass := "test"
	test4 := struct {
		Inner  *Inner
		FieldE string `validate:"omitempty" json:"field_e"`
		Field1 string `validate:"excluded_unless=Inner.Field test" json:"field_1"`
	}{
		Inner:  &Inner{Field: &shouldPass},
		Field1: "filled",
	}
	errs = validate.Struct(test4)
	Equal(t, errs, nil)

	// test7 and test8: excluded_unless has no effect if FieldER is left blank
	test7 := struct {
		Inner   *Inner
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_unless=Inner.Field test" json:"field_er"`
	}{
		FieldE: "test",
	}
	errs = validate.Struct(test7)
	Equal(t, errs, nil)

	test8 := struct {
		FieldE  string `validate:"omitempty" json:"field_e"`
		FieldER string `validate:"excluded_unless=Inner.Field test" json:"field_er"`
	}{
		FieldE: "test",
	}
	errs = validate.Struct(test8)
	Equal(t, errs, nil)

	// Checks number of params in struct tag is correct
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("panicTest should have panicked!")
		}
	}()
	panicTest := struct {
		Inner  *Inner
		Field1 string `validate:"excluded_unless=Inner.Field" json:"field_1"`
	}{
		Inner: &Inner{Field: &fieldVal},
	}
	_ = validate.Struct(panicTest)
}

func Test_hostnameport_validator(t *testing.T) {
	type Host struct {
		Addr string `validate:"hostname_port"`
	}

	type testInput struct {
		data     string
		expected bool
	}
	testData := []testInput{
		{"bad..domain.name:234", false},
		{"extra.dot.com.", false},
		{"localhost:1234", true},
		{"192.168.1.1:1234", true},
		{":1234", true},
		{"domain.com:1334", true},
		{"this.domain.com:234", true},
		{"domain:75000", false},
		{"missing.port", false},
	}
	for _, td := range testData {
		h := Host{Addr: td.data}
		v := New()
		err := v.Struct(h)
		if td.expected != (err == nil) {
			t.Fatalf("Test failed for data: %v Error: %v", td.data, err)
		}
	}
}

func Test_port_validator(t *testing.T) {
	type Host struct {
		Port uint32 `validate:"port"`
	}

	type testInput struct {
		data     uint32
		expected bool
	}

	testData := []testInput{
		{0, false},
		{1, true},
		{65535, true},
		{65536, false},
		{65538, false},
	}
	for _, td := range testData {
		h := Host{Port: td.data}
		v := New()
		err := v.Struct(h)
		if td.expected != (err == nil) {
			t.Fatalf("Test failed for data: %v Error: %v", td.data, err)
		}
	}
}

func TestDurationType(t *testing.T) {
	tests := []struct {
		name    string
		s       interface{} // struct
		success bool
	}{
		{
			name: "valid duration string pass",
			s: struct {
				Value time.Duration `validate:"gte=500ns"`
			}{
				Value: time.Second,
			},
			success: true,
		},
		{
			name: "valid duration int pass",
			s: struct {
				Value time.Duration `validate:"gte=500"`
			}{
				Value: time.Second,
			},
			success: true,
		},
	}

	validate := New()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			errs := validate.Struct(tc.s)
			if tc.success {
				Equal(t, errs, nil)
				return
			}

			NotEqual(t, errs, nil)
		})
	}
}

func TestMultiOrOperatorGroup(t *testing.T) {
	tests := []struct {
		Value    int `validate:"eq=1|gte=5,eq=1|lt=7"`
		expected bool
	}{
		{1, true}, {2, false}, {5, true}, {6, true}, {8, false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Struct(test)
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf("Index: %d multi_group_of_OR_operators failed Error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf("Index: %d multi_group_of_OR_operators should have errs", i)
			}
		}
	}
}

func TestCronExpressionValidation(t *testing.T) {
	tests := []struct {
		value    string `validate:"cron"`
		tag      string
		expected bool
	}{
		{"0 0 12 * * ?", "cron", true},
		{"0 15 10 ? * *", "cron", true},
		{"0 15 10 * * ?", "cron", true},
		{"0 15 10 * * ? 2005", "cron", true},
		{"0 15 10 ? * 6L", "cron", true},
		{"0 15 10 ? * 6L 2002-2005", "cron", true},
		{"*/20 * * * *", "cron", true},
		{"0 15 10 ? * MON-FRI", "cron", true},
		{"0 15 10 ? * 6#3", "cron", true},
		{"0 */15 * * *", "cron", true},
		{"wrong", "cron", false},
	}

	validate := New()
	for i, test := range tests {
		errs := validate.Var(test.value, test.tag)
		if test.expected {
			if !IsEqual(errs, nil) {
				t.Fatalf(`Index: %d cron "%s" failed Error: %s`, i, test.value, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Fatalf(`Index: %d cron "%s" should have errs`, i, test.value)
			}
		}
	}
}

func TestNestedStructValidation(t *testing.T) {
	validator := New(WithRequiredStructEnabled())
	t.Run("required", func(t *testing.T) {
		type (
			value struct {
				Field string
			}
			topLevel struct {
				Nested value `validate:"required"`
			}
		)

		var validationErrs ValidationErrors
		if errs := validator.Struct(topLevel{}); errs != nil {
			validationErrs = errs.(ValidationErrors)
		}

		Equal(t, 1, len(validationErrs))
		AssertError(t, validationErrs, "topLevel.Nested", "topLevel.Nested", "Nested", "Nested", "required")

		Equal(t, validator.Struct(topLevel{value{"potato"}}), nil)
	})

	t.Run("omitempty", func(t *testing.T) {
		type (
			value struct {
				Field string
			}
			topLevel struct {
				Nested value `validate:"omitempty,required"`
			}
		)

		errs := validator.Struct(topLevel{})
		Equal(t, errs, nil)
	})

	t.Run("excluded_if", func(t *testing.T) {
		type (
			value struct {
				Field string
			}
			topLevel struct {
				Field  string
				Nested value `validate:"excluded_if=Field potato"`
			}
		)

		errs := validator.Struct(topLevel{Field: "test", Nested: value{"potato"}})
		Equal(t, errs, nil)

		errs = validator.Struct(topLevel{Field: "potato"})
		Equal(t, errs, nil)

		errs = validator.Struct(topLevel{Field: "potato", Nested: value{"potato"}})
		AssertError(t, errs, "topLevel.Nested", "topLevel.Nested", "Nested", "Nested", "excluded_if")
	})

	t.Run("excluded_unless", func(t *testing.T) {
		type (
			value struct {
				Field string
			}
			topLevel struct {
				Field  string
				Nested value `validate:"excluded_unless=Field potato"`
			}
		)

		errs := validator.Struct(topLevel{Field: "test"})
		Equal(t, errs, nil)

		errs = validator.Struct(topLevel{Field: "potato", Nested: value{"potato"}})
		Equal(t, errs, nil)

		errs = validator.Struct(topLevel{Field: "test", Nested: value{"potato"}})
		AssertError(t, errs, "topLevel.Nested", "topLevel.Nested", "Nested", "Nested", "excluded_unless")
	})

	t.Run("nonComparableField", func(t *testing.T) {
		type (
			value struct {
				Field []string
			}
			topLevel struct {
				Nested value `validate:"required"`
			}
		)

		errs := validator.Struct(topLevel{value{[]string{}}})
		Equal(t, errs, nil)
	})

	type (
		veggyBasket struct {
			Root   string
			Squash string `validate:"required"`
		}
		testErr struct {
			path string
			tag  string
		}
		test struct {
			name  string
			err   testErr
			value veggyBasket
		}
	)

	if err := validator.RegisterValidation("veggy", func(f FieldLevel) bool {
		v, ok := f.Field().Interface().(veggyBasket)
		if !ok || v.Root != "potato" {
			return false
		}
		return true
	}); err != nil {
		t.Fatal(fmt.Errorf("failed to register potato tag: %w", err))
	}

	tests := []test{
		{
			name:  "valid",
			value: veggyBasket{"potato", "zucchini"},
		}, {
			name:  "failedCustomTag",
			value: veggyBasket{"zucchini", "potato"},
			err:   testErr{"topLevel.VeggyBasket", "veggy"},
		}, {
			name:  "failedInnerField",
			value: veggyBasket{"potato", ""},
			err:   testErr{"topLevel.VeggyBasket.Squash", "required"},
		}, {
			name:  "customTagFailurePriorityCheck",
			value: veggyBasket{"zucchini", ""},
			err:   testErr{"topLevel.VeggyBasket", "veggy"},
		},
	}

	evaluateTest := func(tt test, errs error) {
		if tt.err != (testErr{}) && errs != nil {
			Equal(t, len(errs.(ValidationErrors)), 1)

			segments := strings.Split(tt.err.path, ".")
			fieldName := segments[len(segments)-1]
			AssertError(t, errs, tt.err.path, tt.err.path, fieldName, fieldName, tt.err.tag)
		}

		shouldFail := tt.err != (testErr{})
		hasFailed := errs != nil
		if shouldFail != hasFailed {
			t.Fatalf("expected failure %v, got: %v with errs: %v", shouldFail, hasFailed, errs)
		}
	}

	for _, tt := range tests {
		type topLevel struct {
			VeggyBasket veggyBasket `validate:"veggy"`
		}

		t.Run(tt.name, func(t *testing.T) {
			evaluateTest(tt, validator.Struct(topLevel{tt.value}))
		})
	}

	// Also test on struct pointers
	for _, tt := range tests {
		type topLevel struct {
			VeggyBasket *veggyBasket `validate:"veggy"`
		}

		t.Run(tt.name+"Ptr", func(t *testing.T) {
			evaluateTest(tt, validator.Struct(topLevel{&tt.value}))
		})
	}
}

func TestTimeRequired(t *testing.T) {
	validate := New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "-" {
			return name
		}

		return ""
	})

	type TestTime struct {
		Time time.Time `validate:"required"`
	}

	var testTime TestTime
	err := validate.Struct(&testTime)
	NotEqual(t, err, nil)
	AssertError(t, err.(ValidationErrors), "TestTime.Time", "TestTime.Time", "Time", "Time", "required")
}

func TestOmitNilAndRequired(t *testing.T) {
	type (
		OmitEmpty struct {
			Str    string  `validate:"omitempty,required,min=10"`
			StrPtr *string `validate:"omitempty,required,min=10"`
			Inner  *OmitEmpty
		}
		OmitNil struct {
			Str    string  `validate:"omitnil,required,min=10"`
			StrPtr *string `validate:"omitnil,required,min=10"`
			Inner  *OmitNil
		}
	)

	var (
		validate = New(WithRequiredStructEnabled())
		valid    = "this is the long string to pass the validation rule"
	)

	t.Run("compare using valid data", func(t *testing.T) {
		err1 := validate.Struct(OmitEmpty{Str: valid, StrPtr: &valid, Inner: &OmitEmpty{Str: valid, StrPtr: &valid}})
		err2 := validate.Struct(OmitNil{Str: valid, StrPtr: &valid, Inner: &OmitNil{Str: valid, StrPtr: &valid}})

		Equal(t, err1, nil)
		Equal(t, err2, nil)
	})

	t.Run("compare fully empty omitempty and omitnil", func(t *testing.T) {
		err1 := validate.Struct(OmitEmpty{})
		err2 := validate.Struct(OmitNil{})

		Equal(t, err1, nil)
		AssertError(t, err2, "OmitNil.Str", "OmitNil.Str", "Str", "Str", "required")
	})

	t.Run("validate in deep", func(t *testing.T) {
		err1 := validate.Struct(OmitEmpty{Str: valid, Inner: &OmitEmpty{}})
		err2 := validate.Struct(OmitNil{Str: valid, Inner: &OmitNil{}})

		Equal(t, err1, nil)
		AssertError(t, err2, "OmitNil.Inner.Str", "OmitNil.Inner.Str", "Str", "Str", "required")
	})
}

func TestOmitZero(t *testing.T) {
	type (
		OmitEmpty struct {
			Str    string  `validate:"omitempty,min=10"`
			StrPtr *string `validate:"omitempty,min=10"`
		}
		OmitZero struct {
			Str    string  `validate:"omitzero,min=10"`
			StrPtr *string `validate:"omitzero,min=10"`
		}
	)

	var (
		validate = New()
		valid    = "this is the long string to pass the validation rule"
		empty    = ""
	)

	t.Run("compare using valid data", func(t *testing.T) {
		err1 := validate.Struct(OmitEmpty{Str: valid, StrPtr: &valid})
		err2 := validate.Struct(OmitZero{Str: valid, StrPtr: &valid})

		Equal(t, err1, nil)
		Equal(t, err2, nil)
	})

	t.Run("compare fully empty omitempty and omitzero", func(t *testing.T) {
		err1 := validate.Struct(OmitEmpty{})
		err2 := validate.Struct(OmitZero{})

		Equal(t, err1, nil)
		Equal(t, err2, nil)
	})

	t.Run("compare with zero value", func(t *testing.T) {
		err1 := validate.Struct(OmitEmpty{Str: "", StrPtr: nil})
		err2 := validate.Struct(OmitZero{Str: "", StrPtr: nil})

		Equal(t, err1, nil)
		Equal(t, err2, nil)
	})

	t.Run("compare with empty value", func(t *testing.T) {
		err1 := validate.Struct(OmitEmpty{Str: empty, StrPtr: &empty})
		err2 := validate.Struct(OmitZero{Str: empty, StrPtr: &empty})

		AssertError(t, err1, "OmitEmpty.StrPtr", "OmitEmpty.StrPtr", "StrPtr", "StrPtr", "min")
		Equal(t, err2, nil)
	})
}

func AssertError(t *testing.T, err error, nsKey, structNsKey, field, structField, expectedTag string) {
	var found bool
	var fe FieldError
	errs := err.(ValidationErrors)
	for i := 0; i < len(errs); i++ {
		if errs[i].Namespace() == nsKey && errs[i].StructNamespace() == structNsKey {
			found = true
			fe = errs[i]
			break
		}
	}

	EqualSkip(t, 2, found, true)
	NotEqualSkip(t, 2, fe, nil)
	EqualSkip(t, 2, fe.Field(), field)
	EqualSkip(t, 2, fe.StructField(), structField)
	EqualSkip(t, 2, fe.Tag(), expectedTag)
}

func AssertDeepError(t *testing.T, err error, nsKey, structNsKey, field, structField, expectedTag, actualTag string) {
	var found bool
	var fe FieldError
	errs := err.(ValidationErrors)
	for i := 0; i < len(errs); i++ {
		if errs[i].Namespace() == nsKey && errs[i].StructNamespace() == structNsKey && errs[i].Tag() == expectedTag && errs[i].ActualTag() == actualTag {
			found = true
			fe = errs[i]
			break
		}
	}

	EqualSkip(t, 2, found, true)
	NotEqualSkip(t, 2, fe, nil)
	EqualSkip(t, 2, fe.Field(), field)
	EqualSkip(t, 2, fe.StructField(), structField)
}

func getError(err error, nsKey, structNsKey string) (fe FieldError) {
	errs := err.(ValidationErrors)
	for i := 0; i < len(errs); i++ {
		if errs[i].Namespace() == nsKey && errs[i].StructNamespace() == structNsKey {
			fe = errs[i]
			break
		}
	}

	return
}
