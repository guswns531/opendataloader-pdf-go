// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0

package processors

import (
	"sort"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/containers"
	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const (
	clusterRowTolerance = 6.0
	clusterGapTolerance = 24.0
)

type ClusterTableProcessor struct {
	AbstractTableProcessor
}

func (p *ClusterTableProcessor) Process(elements []entities.IObject, ctx *containers.ProcessorContext) []entities.IObject {
	chunks := make([]*entities.TextChunk, 0)
	for _, element := range elements {
		chunk, ok := element.(*entities.TextChunk)
		if !ok || isWhitespaceChunk(chunk) {
			continue
		}
		chunks = append(chunks, chunk)
	}
	if len(chunks) < 4 {
		return append([]entities.IObject(nil), elements...)
	}

	clusters := dbscanTextClusters(chunks)
	if len(clusters) == 0 {
		return append([]entities.IObject(nil), elements...)
	}

	used := make(map[string]struct{})
	result := make([]entities.IObject, 0, len(elements))
	for _, cluster := range clusters {
		table := clusterToTable(cluster, ctx)
		if table == nil || len(table.Rows) < 2 || len(table.Rows[0].Cells) < 2 {
			continue
		}
		result = append(result, table)
		for _, chunk := range cluster {
			used[chunk.GetID()] = struct{}{}
		}
	}
	for _, element := range elements {
		if _, ok := used[element.GetID()]; !ok {
			result = append(result, element)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		left := result[i].GetBBox()
		right := result[j].GetBBox()
		if !areClose(left.Y, right.Y, tableAlignmentTolerance) {
			return left.Y > right.Y
		}
		return left.X < right.X
	})
	return result
}

func dbscanTextClusters(chunks []*entities.TextChunk) [][]*entities.TextChunk {
	visited := make([]bool, len(chunks))
	var clusters [][]*entities.TextChunk

	for idx := range chunks {
		if visited[idx] {
			continue
		}
		visited[idx] = true
		neighbors := clusterNeighbors(chunks, idx)
		if len(neighbors) < 3 {
			continue
		}
		cluster := []*entities.TextChunk{chunks[idx]}
		queue := append([]int(nil), neighbors...)
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			if !visited[current] {
				visited[current] = true
				more := clusterNeighbors(chunks, current)
				if len(more) >= 3 {
					queue = append(queue, more...)
				}
			}
			if !containsChunk(cluster, chunks[current]) {
				cluster = append(cluster, chunks[current])
			}
		}
		clusters = append(clusters, cluster)
	}
	return clusters
}

func clusterNeighbors(chunks []*entities.TextChunk, index int) []int {
	neighbors := make([]int, 0)
	source := chunks[index]
	for idx, chunk := range chunks {
		if idx == index || source.GetBBox().Page != chunk.GetBBox().Page {
			continue
		}
		if areClose(source.Baseline, chunk.Baseline, clusterRowTolerance) ||
			(source.GetBBox().X <= chunk.GetBBox().X && chunk.GetBBox().X-bboxRight(source.GetBBox()) <= clusterGapTolerance &&
				areClose(source.GetBBox().Y, chunk.GetBBox().Y, clusterRowTolerance)) {
			neighbors = append(neighbors, idx)
		}
	}
	return neighbors
}

func containsChunk(cluster []*entities.TextChunk, chunk *entities.TextChunk) bool {
	for _, item := range cluster {
		if item == chunk {
			return true
		}
	}
	return false
}

func clusterToTable(cluster []*entities.TextChunk, ctx *containers.ProcessorContext) *entities.SemanticTable {
	sort.SliceStable(cluster, func(i, j int) bool {
		if !areClose(cluster[i].Baseline, cluster[j].Baseline, clusterRowTolerance) {
			return cluster[i].Baseline > cluster[j].Baseline
		}
		return cluster[i].GetBBox().X < cluster[j].GetBBox().X
	})

	rowGroups := make([][]*entities.TextChunk, 0)
	for _, chunk := range cluster {
		if len(rowGroups) == 0 || !areClose(rowGroups[len(rowGroups)-1][0].Baseline, chunk.Baseline, clusterRowTolerance) {
			rowGroups = append(rowGroups, []*entities.TextChunk{chunk})
			continue
		}
		rowGroups[len(rowGroups)-1] = append(rowGroups[len(rowGroups)-1], chunk)
	}
	if len(rowGroups) < 2 {
		return nil
	}

	colStarts := make([]float64, 0)
	cells := make(map[[2]int][]entities.IObject)
	rowBounds := make([]float64, 0, len(rowGroups)+1)
	page := rowGroups[0][0].GetBBox().Page
	maxRight := 0.0
	for rowIdx, row := range rowGroups {
		sort.SliceStable(row, func(i, j int) bool { return row[i].GetBBox().X < row[j].GetBBox().X })
		if len(row) < 2 {
			return nil
		}
		top := bboxTop(row[0].GetBBox())
		bottom := row[0].GetBBox().Y
		for _, chunk := range row {
			top = max(top, bboxTop(chunk.GetBBox()))
			bottom = min(bottom, chunk.GetBBox().Y)
			colStarts = append(colStarts, chunk.GetBBox().X)
			maxRight = max(maxRight, bboxRight(chunk.GetBBox()))
		}
		if rowIdx == 0 {
			rowBounds = append(rowBounds, top)
		}
		rowBounds = append(rowBounds, bottom)
	}

	colBounds := uniqueSorted(append(colStarts, maxRight), tableAlignmentTolerance, false)
	if len(colBounds) < 3 {
		return nil
	}
	for rowIdx, row := range rowGroups {
		for _, chunk := range row {
			col := nearestColumn(colBounds, chunk.GetBBox().X)
			if col < 0 {
				continue
			}
			cells[[2]int{rowIdx, col}] = append(cells[[2]int{rowIdx, col}], chunk)
		}
	}

	return buildTableFromGrid(rowBounds, colBounds, cells, page, ctx)
}
