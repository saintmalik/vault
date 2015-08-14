package ssh

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

type configLease struct {
	Lease    time.Duration
	LeaseMax time.Duration
}

func pathConfigLease(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config/lease",
		Fields: map[string]*framework.FieldSchema{
			"lease": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "[Required] Default lease for roles.",
			},
			"lease_max": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "[Required] Maximum time a credential is valid for.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.WriteOperation: b.pathConfigLeaseWrite,
		},

		HelpSynopsis:    pathConfigLeaseHelpSyn,
		HelpDescription: pathConfigLeaseHelpDesc,
	}
}

func (b *backend) pathConfigLeaseWrite(req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	leaseRaw := d.Get("lease").(string)
	if leaseRaw == "" {
		return logical.ErrorResponse("Missing lease"), nil
	}

	leaseMaxRaw := d.Get("lease_max").(string)
	if leaseMaxRaw == "" {
		return logical.ErrorResponse("Missing lease_max"), nil
	}

	lease, err := time.ParseDuration(leaseRaw)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf(
			"Invalid 'lease': %s", err)), nil
	}

	leaseMax, err := time.ParseDuration(leaseMaxRaw)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf(
			"Invalid 'lease_max': %s", err)), nil
	}

	entry, err := logical.StorageEntryJSON("config/lease", &configLease{
		Lease:    lease,
		LeaseMax: leaseMax,
	})

	if err != nil {
		return nil, fmt.Errorf("could not create storage entry JSON: %s", err)
	}

	if err := req.Storage.Put(entry); err != nil {
		return nil, fmt.Errorf("could not store JSON: %s", err)
	}

	return nil, nil
}

func (b *backend) Lease(s logical.Storage) (*configLease, error) {
	entry, err := s.Get("config/lease")

	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	var result configLease
	if err := entry.DecodeJSON(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

const pathConfigLeaseHelpSyn = `
Configure the default lease information for SSH dynamic keys.
`

const pathConfigLeaseHelpDesc = `
This configures the default lease information used for SSH keys 
generated by this backend. The lease specifies the duration that a
credential will be valid for, as well as the maximum session for
a set of credentials.

The format for the lease is "1h" or integer and then unit. The longest
unit is hour.
`
