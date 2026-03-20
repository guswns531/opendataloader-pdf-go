package core

import "io"

type Ingestor interface {
	Name() string
	Ingest(ctx *ProcessingContext, source Source) (*Document, error)
}

type Heuristic interface {
	Name() string
	Apply(ctx *ProcessingContext, document *Document) error
}

type Emitter interface {
	Name() string
	Format() OutputFormat
	Emit(ctx *ProcessingContext, document *Document, w io.Writer) error
}

type Processor interface {
	Name() string
	Process(ctx *ProcessingContext, document *Document) error
}

type Pipeline interface {
	Run(ctx *ProcessingContext, source Source, emitter Emitter) (*Document, error)
}

type HeuristicSet []Heuristic

func (set HeuristicSet) Apply(ctx *ProcessingContext, document *Document) error {
	for _, heuristic := range set {
		if err := heuristic.Apply(ctx, document); err != nil {
			return err
		}
	}
	return nil
}
