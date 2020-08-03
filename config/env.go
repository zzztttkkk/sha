package config

func (t *Type) IsDebug() bool { return t.isDebug }

func (t *Type) IsRelease() bool { return t.isRelease }

func (t *Type) IsTest() bool { return t.isTest }
