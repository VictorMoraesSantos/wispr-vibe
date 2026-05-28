package processor

type Filter func(string) string

type Pipeline struct {
	filters []Filter
}

func NewPipeline(filters ...Filter) *Pipeline {
	return &Pipeline{filters: filters}
}

func (p *Pipeline) Process(text string) string {
	for _, f := range p.filters {
		text = f(text)
	}
	return text
}

func (p *Pipeline) Add(f Filter) {
	p.filters = append(p.filters, f)
}

func DefaultPipeline() *Pipeline {
	return NewPipeline(
		RemoveFillers,
		CollapseSpaces,
		TrimText,
		CapitalizeFirst,
	)
}
