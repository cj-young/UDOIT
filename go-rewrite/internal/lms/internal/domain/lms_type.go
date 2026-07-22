package domain

type LMSType string

const (
	LMSTypeCanvas LMSType = "canvas"
	LMSTypeD2L	LMSType = "d2l"
)

func (t LMSType) IsValid() bool {
	switch t {
	case LMSTypeCanvas, LMSTypeD2L:
		return true
	default:
		return false
	}
}