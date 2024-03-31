package bus

import (
	"errors"
	"fmt"
	"testing"
)

type testQuery struct {
	Id   int64
	Resp string
}

func TestEventListeners(t *testing.T) {
	count := 0

	// EventLister 支持多个，顺序执行，会阻塞
	err := Publish(&testQuery{})
	fmt.Println("publish before add eventListener", err)
	AddEventListener(func(query *testQuery) error {
		count += 1
		fmt.Println(count)
		//time.Sleep(3 * time.Second)
		return nil
	})

	AddEventListener(func(query *testQuery) error {
		count += 10
		fmt.Println(count)
		return nil
	})

	err = Publish(&testQuery{})

	if err != nil {
		t.Fatal("Publish event failed ", err.Error())
	} else if count != 11 {
		t.Fatal(fmt.Sprintf("Publish event failed, listeners called: %v, expected: %v", count, 11))
	}
	fmt.Println(count)

	// 只支持一个，后注册的会覆盖前面的， Only support on handler, will overwrite.
	err = Dispatch(&testQuery{
		Id: 55,
	})
	fmt.Println("dispatch before add handler", err)
	AddHandler("", func(query *testQuery) error {
		fmt.Println("handler", query.Id)
		return errors.New("dd")
	})
	AddHandler("", func(query *testQuery) error {
		fmt.Println("handler2", query.Id)
		return errors.New("dd")
	})
	err = Dispatch(&testQuery{
		Id: 55,
	})
	fmt.Println(err)
}
