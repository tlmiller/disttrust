package vault

import (
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/hashicorp/vault/api"

	"github.com/pkg/errors"

	"github.com/tlmiller/disttrust/provider"
)

type Lease struct {
	leaseID   string
	renewable bool
	request   *provider.Request
	response  *provider.Response
	start     time.Time
	till      time.Time
}

func (l *Lease) HasResponse() bool {
	return l.response != nil
}

func (l *Lease) ID() string {
	return l.leaseID
}

func LeaseFromSecret(req *provider.Request, secret *api.Secret) (*Lease, error) {
	lease := Lease{}

	lease.start = time.Now()
	lease.renewable = secret.Renewable
	if lease.renewable {
		lease.leaseID = secret.LeaseID
	} else {
		lease.leaseID = secret.RequestID
	}

	res, err := makeResponse(secret.Data)
	if err != nil {
		return nil, errors.Wrap(err, "making lease response")
	}
	lease.response = res

	if secret.LeaseDuration != 0 {
		lease.till = time.Now().Add(time.Duration(secret.LeaseDuration) * time.Second)
	} else {
		block, _ := pem.Decode([]byte(res.Certificate))
		if block == nil {
			return nil, errors.Wrap(err, "decoding pem lease certificate")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "making lease renew time")
		}
		lease.till = cert.NotAfter
	}

	return &lease, nil
}

func makeResponse(data map[string]interface{}) (*provider.Response, error) {
	res := provider.Response{}
	var ok bool
	res.Certificate, ok = data["certificate"].(string)
	if !ok {
		return nil, errors.New("unknown type for issued certificate")
	}
	res.PrivateKey, ok = data["private_key"].(string)
	if !ok {
		return nil, errors.New("unknown type for issued private key")
	}
	res.Serial, ok = data["serial_number"].(string)
	if !ok {
		return nil, errors.New("unknown type for issued serial")
	}
	return &res, nil
}

func (l *Lease) Request() *provider.Request {
	return l.request
}

func (l *Lease) Response() (*provider.Response, error) {
	return l.response, nil
}

func (l *Lease) Start() time.Time {
	return l.start
}

func (l *Lease) Till() time.Time {
	return l.till
}
