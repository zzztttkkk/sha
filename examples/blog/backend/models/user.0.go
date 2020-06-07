package models

type User struct {
	model
	Name     string `json:"name" ddl:":notnull;unique;L<30>"`
	Alias    string `json:"alias" ddl:":L<30>;D<''>"`
	Password []byte `json:"-" ddl:":notnull;L<64>"`
	Bio      string `json:"bio" ddl:":L<120>;D<''>"`
	Avatar   string `json:"avatar" ddl:":L<120>;D<''>"`
}
