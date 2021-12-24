package mock

import (
	"errors"
	"github.com/golang/mock/gomock"
	"testing"
)

//Return 返回确定的值
//Do Mock 方法被调用时，要执行的操作吗，忽略返回值。
//DoAndReturn 可以动态地控制返回值
func TestGetFromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // 断言 DB.Get() 方法是否被调用

	m := NewMockDB(ctrl)
	m.EXPECT().Get(gomock.Eq("Tom")).Return(100, errors.New("not exist"))

	m.EXPECT().Get(gomock.Not("Sam")).Return(0, nil)
	m.EXPECT().Get(gomock.Any()).Do(func(key string) {
		t.Log(key)
	})
	m.EXPECT().Get(gomock.Any()).DoAndReturn(func(key string) (int, error) {
		if key == "Sam" {
			return 630, nil
		}
		return 0, errors.New("not exist")
	})

	if v := GetFromDB(m, "Tom"); v != -1 {
		t.Fatal("expected -1, but got", v)
	}
}

//Times() 断言 Mock 方法被调用的次数。
//MaxTimes() 最大次数。
//MinTimes() 最小次数。
//AnyTimes() 任意次数（包括 0 次）。
func TestGetFromDB2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockDB(ctrl)
	m.EXPECT().Get(gomock.Not("Sam")).Return(0, nil).Times(2)
	GetFromDB(m, "ABC")
	GetFromDB(m, "DEF")
}

//调用顺序
func TestGetFromDB3(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // 断言 DB.Get() 方法是否被调用

	m := NewMockDB(ctrl)
	o1 := m.EXPECT().Get(gomock.Eq("Tom")).Return(0, errors.New("not exist"))
	o2 := m.EXPECT().Get(gomock.Eq("Sam")).Return(630, nil)
	gomock.InOrder(o1, o2)
	GetFromDB(m, "Tom")
	GetFromDB(m, "Sam")
}
