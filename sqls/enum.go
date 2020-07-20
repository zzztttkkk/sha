package sqls

type Enumer interface {
	GetId() int64
	GetName() string
}

type Enum struct {
	Model
	Name string `json:"name"`
}

func (enum *Enum) GetId() int64 {
	return enum.Id
}

func (enum *Enum) GetName() string {
	return enum.Name
}
