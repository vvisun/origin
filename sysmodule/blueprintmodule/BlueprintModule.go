package blueprintmodule

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/blueprint"
	"sync/atomic"
)

type BlueprintModule struct {
	service.Module

	bp blueprint.Blueprint

	execDefFilePath string
	graphFilePath   string

	seedGraphID int64

	mapGraph map[int64]blueprint.IGraph
}

func (m *BlueprintModule) Init(execDefFilePath string, graphFilePath string) error {
	m.execDefFilePath = execDefFilePath
	m.graphFilePath = graphFilePath

	m.mapGraph = make(map[int64]blueprint.IGraph, 1024)
	return nil
}

func (m *BlueprintModule) OnInit() error {
	if m.execDefFilePath == "" || m.graphFilePath == "" {
		return fmt.Errorf("execDefFilePath or graphFilePath is empty")
	}

	m.seedGraphID = 1
	return m.bp.Init(m.execDefFilePath, m.graphFilePath, m)
}

func (m *BlueprintModule) CreateGraph(graphName string) int64 {
	graphID := atomic.AddInt64(&m.seedGraphID, 1)
	graph := m.bp.Create(graphName, graphID)
	if graph == nil {
		return 0
	}
	m.mapGraph[graphID] = graph

	return graphID
}

func (m *BlueprintModule) GetGraph(graphID int64) (blueprint.IGraph, error) {
	graph, ok := m.mapGraph[graphID]
	if !ok {
		return nil, fmt.Errorf("graph not found,graphID:%d", graphID)
	}
	return graph, nil
}

func (m *BlueprintModule) Do(graphID int64, entranceID int64, args ...any) error {
	graph, err := m.GetGraph(graphID)
	if err != nil {
		return err
	}
	return graph.Do(entranceID, args...)
}

func (m *BlueprintModule) TriggerEvent(graphID int64, eventID int64, args ...any) error {
	graph, err := m.GetGraph(graphID)
	if err != nil {
		return err
	}

	return graph.Do(eventID, args...)
}
