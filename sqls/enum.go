package sqls

type EnumItem interface {
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

func (enum Enum) TableDefinition(lines ...string) []string {
	return enum.Model.TableDefinition(append(lines, "name char(255) not null unique")...)
}
