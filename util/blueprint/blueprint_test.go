package blueprint

import (
	"testing"
)

func TestExecMgr(t *testing.T) {
	var bp Blueprint
	err := bp.Init("E:\\WorkSpace\\c4\\OriginNodeEditor\\json", "E:\\WorkSpace\\c4\\OriginNodeEditor\\vgf", nil, nil)
	if err != nil {
		t.Fatalf("Init failed,err:%v", err)
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
