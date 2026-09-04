package extensions

type SupportedExtension string

const (
	Png SupportedExtension = ".png"
)

func (a SupportedExtension) IsValid() bool {
	switch a {
	case Png:
		return true
	default:
		return false
	}
}
