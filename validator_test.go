package validator

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
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
