// Copyright 2025-2026 Hancom Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package containers

import (
	"strconv"
	"sync/atomic"
)

type ProcessorContext struct {
	idCounter        int64
	Headings         []interface{}
	ImageIndex       int
	UseStructTree    bool
	ContrastConsumer func(float64)
	ImageDir         string
	EmbedImages      bool
	ImageFormat      string
}

func NewProcessorContext() *ProcessorContext {
	return &ProcessorContext{
		Headings: make([]interface{}, 0),
	}
}

func (c *ProcessorContext) NextID() string {
	next := atomic.AddInt64(&c.idCounter, 1)
	return strconv.FormatInt(next, 10)
}
