package blueprint

import (
	"fmt"
)

type Blueprint struct {
	execPool  ExecPool
	graphPool GraphPool
}

func (bm *Blueprint) Init(execDefFilePath string, graphFilePath string) error {
	err := bm.execPool.Load(execDefFilePath)
	if err != nil {
		return err
	}
	
	for _, e := range execNodes {
		if !bm.execPool.Register(e) {
			return fmt.Errorf("register exec failed,exec:%s", e.GetName())
		}
	}

	err = bm.graphPool.Load(&bm.execPool, graphFilePath)
	if err != nil {
		return err
	}

	return nil
}

func (bm *Blueprint) Create(graphName string) IGraph {
	return bm.graphPool.Create(graphName)
}
