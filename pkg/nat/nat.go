package nat

type Nat int

func (t Nat) String() string {
	switch t {
	case EIM:
		return "EIM"
	case EDM:
		return "EDM"
	default:
		return ""
	}
}

const (
	EIM Nat = iota + 1
	EDM
)
