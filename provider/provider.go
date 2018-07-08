package provider

type Id string

type Provider interface {
	Issue(*Request) (Lease, error)
}
