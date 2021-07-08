package internal

type Status16 uint16

func (s *Status16) Add(v uint8) { *s |= 1 << v }

func (s *Status16) Del(v uint8) { *s &= ^(1 << v) }

func (s *Status16) Has(v uint8) bool { return (*s >> v & 1) != 0 }

func (s *Status16) Reset() { *s = 0 }
