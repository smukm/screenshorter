package service

type Screenshot interface {
	Make() string
}

type Service struct {
	Screenshot Screenshot
}

func NewService(s Screenshot) *Service {

	return &Service{
		Screenshot: s,
	}
}
