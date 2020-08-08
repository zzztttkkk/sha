package config

func (t *Config) IsDebug() bool { return t.isDebug }

func (t *Config) IsRelease() bool { return t.isRelease }

func (t *Config) IsTest() bool { return t.isTest }
