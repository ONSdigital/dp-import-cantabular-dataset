package processor

type Processor struct{
	numWorkers int
}

func New(numWorkers int) *Processor{
	return &Processor{
		numWorkers: numWorkers,
	}
}