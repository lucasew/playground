package plugin

type Plugin interface {
	GetLookuper(namespace []string) Lookuper
}
