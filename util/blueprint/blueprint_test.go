package blueprint

import (
	"testing"
)

func TestExecMgr(t *testing.T) {
	var bp Blueprint
	err := bp.Init("D:\\Develop\\OriginNodeEditor\\json", "D:\\Develop\\OriginNodeEditor\\vgf")
	if err != nil {
		t.Fatalf("init failed,err:%v", err)
	}

	graphTest1 := bp.Create("testSwitch")
	err = graphTest1.Do(EntranceID_IntParam, 2, 1, 3)
	if err != nil {
		t.Fatalf("Do EntranceID_IntParam failed,err:%v", err)
	}

	//graphTest2 := bp.Create("testForeach")
	//err = graphTest2.Do(EntranceID_IntParam, 1, 2, 3)
	//if err != nil {
	//	t.Fatalf("Do EntranceID_IntParam failed,err:%v", err)
	//}

	//graphTest2 := bp.Create("test2")
	//
	//err = graphTest2.Do(EntranceID_IntParam, 1, 2, 3)
	//if err != nil {
	//	t.Fatalf("Do EntranceID_IntParam failed,err:%v", err)
	//}

	//graph := bp.Create("test1")
	//err = graph.Do(EntranceID_IntParam, 1, 2, 3)
	//if err != nil {
	//	t.Fatalf("do failed,err:%v", err)
	//}

	//graph.Release()
}
