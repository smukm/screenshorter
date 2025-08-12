package service

type Playwrite struct {
}

func NewPlaywrite() *Playwrite {
	return &Playwrite{}
}

func (p *Playwrite) Make() string {

	return "screenshot.png"
}
