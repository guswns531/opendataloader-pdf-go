package core

type ProcessingContext struct {
	Document   *Document
	Options    ProcessingOptions
	Config     ProcessingConfig
	Stage      Stage
	Values     map[string]any
	PageValues map[PageNumber]map[string]any
	NodeValues map[ElementID]map[string]any
	Issues     []Issue
}

func NewProcessingContext(document *Document, options ProcessingOptions) *ProcessingContext {
	return &ProcessingContext{
		Document:   document,
		Options:    options,
		Config:     options.Resolve(),
		Values:     make(map[string]any),
		PageValues: make(map[PageNumber]map[string]any),
		NodeValues: make(map[ElementID]map[string]any),
	}
}

func (c *ProcessingContext) SetDocument(document *Document) {
	c.Document = document
}

func (c *ProcessingContext) SetStage(stage Stage) {
	c.Stage = stage
}

func (c *ProcessingContext) Set(key string, value any) {
	if c.Values == nil {
		c.Values = make(map[string]any)
	}
	c.Values[key] = value
}

func (c *ProcessingContext) Get(key string) (any, bool) {
	if c.Values == nil {
		return nil, false
	}
	value, ok := c.Values[key]
	return value, ok
}

func (c *ProcessingContext) SetPageValue(page PageNumber, key string, value any) {
	if c.PageValues == nil {
		c.PageValues = make(map[PageNumber]map[string]any)
	}
	values := c.PageValues[page]
	if values == nil {
		values = make(map[string]any)
		c.PageValues[page] = values
	}
	values[key] = value
}

func (c *ProcessingContext) PageValue(page PageNumber, key string) (any, bool) {
	if c.PageValues == nil {
		return nil, false
	}
	values, ok := c.PageValues[page]
	if !ok {
		return nil, false
	}
	value, ok := values[key]
	return value, ok
}

func (c *ProcessingContext) SetNodeValue(id ElementID, key string, value any) {
	if c.NodeValues == nil {
		c.NodeValues = make(map[ElementID]map[string]any)
	}
	values := c.NodeValues[id]
	if values == nil {
		values = make(map[string]any)
		c.NodeValues[id] = values
	}
	values[key] = value
}

func (c *ProcessingContext) NodeValue(id ElementID, key string) (any, bool) {
	if c.NodeValues == nil {
		return nil, false
	}
	values, ok := c.NodeValues[id]
	if !ok {
		return nil, false
	}
	value, ok := values[key]
	return value, ok
}

func (c *ProcessingContext) AddIssue(issue Issue) {
	c.Issues = append(c.Issues, issue)
}
