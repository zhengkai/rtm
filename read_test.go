package rtm

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {

	data1 := []byte{70, 80, 78, 78, 1, 64, 2, 0, 44, 0, 0, 0, 246, 100, 86, 110, 123, 34, 116, 111, 107, 101, 110, 34, 58, 34, 70, 66, 67, 66, 54, 57, 53, 49, 53, 48, 66, 53, 52, 57, 49, 48, 70, 48, 55, 49, 65, 65, 52, 66, 66, 57, 52, 69, 56, 67, 56, 50, 34, 125}
	data2 := []byte{70, 80, 78, 78, 1, 128, 1, 7, 68, 0, 0, 0, 75, 39, 1, 0, 112, 117, 115, 104, 109, 115, 103, 135, 164, 102, 114, 111, 109, 205, 78, 39, 162, 116, 111, 205, 39, 17, 163, 109, 105, 100, 207, 21, 113, 104, 38, 171, 47, 201, 212, 165, 109, 116, 121, 112, 101, 37, 163, 109, 115, 103, 166, 116, 101, 115, 116, 32, 49, 165, 97, 116, 116, 114, 115, 160, 165, 109, 116, 105, 109, 101, 207, 0, 0, 1, 103, 192, 246, 218, 36}

	fmt.Println(len(data1))

	// 单独测试两个基本数据有没有问题 part 1

	r1, remain, err := readByte(data1)

	if reflect.TypeOf(r1).String() != `[]*rtm.Read` {
		t.Error(`data1 type error`)
	}

	if len(remain) != 0 {
		t.Error(`data1 length error`)
	}

	if err != nil {
		t.Error(`data1 parse error`)
	}

	// 单独测试两个基本数据有没有问题 part 2

	r2, remain, err := readByte(data2)
	if reflect.TypeOf(r2).String() != `[]*rtm.Read` {
		t.Error(`data2 type error`)
	}

	if len(remain) != 0 {
		t.Error(`data2 length error`)
	}

	if err != nil {
		t.Error(`data2 parse error`)
	}

	if reflect.DeepEqual(r1, r2) {
		t.Error(`data1 & data2 are equal`)
	}

	size := len(data1)
	for i := 0; i < size; i++ {

		// 不完整数据
		r, remain, err := readByte(data1[:i])
		if r != nil {
			t.Error(`broken data1 parse error`, i)
		}
		if len(remain) != i {
			t.Error(`broken data1 remain error`, i, len(remain))
		}
		if err != nil {
			t.Error(`broken data1 error`, i)
		}

		// 第一个完整，第二个不完整
		data := make([]byte, size+i)
		copy(data, data1)
		copy(data[size:], data2)

		r, remain, err = readByte(data)
		if r == nil || len(r) != 1 {
			t.Error(`data3 parse error`, i)
		}
		if !reflect.DeepEqual(r, r1) || reflect.DeepEqual(r, r2) {
			t.Error(`data3 parse error`, i)
		}
		if len(remain) != i {
			t.Error(`data3 remain error`, i, len(remain))
		}
		if err != nil {
			t.Error(`data3 error`, i)
		}

		// fmt.Println(`data3`, r, len(remain), err)
	}
}
