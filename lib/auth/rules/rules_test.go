package rules

import (
	"testing"

	"github.com/lack-io/vine/lib/auth"
)

func TestVerify(t *testing.T) {
	svcResource := &auth.Resource{
		Type:     "service",
		Name:     "go.vine.service.foo",
		Endpoint: "Foo.Bar",
	}

	webResource := &auth.Resource{
		Type:     "service",
		Name:     "go.vine.web.foo",
		Endpoint: "/foo/bar",
	}

	catchallResource := &auth.Resource{
		Type:     "*",
		Name:     "*",
		Endpoint: "*",
	}

	tt := []struct {
		Name     string
		Rules    []*auth.Rule
		Account  *auth.Account
		Resource *auth.Resource
		Error    error
	}{
		{
			Name:     "NoRules",
			Rules:    []*auth.Rule{},
			Account:  nil,
			Resource: svcResource,
			Error:    auth.ErrForbidden,
		},
		{
			Name:     "CatchallPublicAccount",
			Account:  &auth.Account{},
			Resource: svcResource,
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "",
					Resource: catchallResource,
				},
			},
		},
		{
			Name:     "CatchallPublicNoAccount",
			Resource: svcResource,
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "",
					Resource: catchallResource,
				},
			},
		},
		{
			Name:     "CatchallPrivateAccount",
			Account:  &auth.Account{},
			Resource: svcResource,
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "*",
					Resource: catchallResource,
				},
			},
		},
		{
			Name:     "CatchallPrivateNoAccount",
			Resource: svcResource,
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "*",
					Resource: catchallResource,
				},
			},
			Error: auth.ErrForbidden,
		},
		{
			Name:     "CatchallServiceRuleMatch",
			Resource: svcResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope: "*",
					Resource: &auth.Resource{
						Type:     svcResource.Type,
						Name:     svcResource.Name,
						Endpoint: "*",
					},
				},
			},
		},
		{
			Name:     "CatchallServiceRuleNoMatch",
			Resource: svcResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope: "*",
					Resource: &auth.Resource{
						Type:     svcResource.Type,
						Name:     "wrongname",
						Endpoint: "*",
					},
				},
			},
			Error: auth.ErrForbidden,
		},
		{
			Name:     "ExactRuleValidScope",
			Resource: svcResource,
			Account: &auth.Account{
				Scopes: []string{"neededscope"},
			},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "neededscope",
					Resource: svcResource,
				},
			},
		},
		{
			Name:     "ExactRuleInvalidScope",
			Resource: svcResource,
			Account: &auth.Account{
				Scopes: []string{"neededscope"},
			},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "invalidscope",
					Resource: svcResource,
				},
			},
			Error: auth.ErrForbidden,
		},
		{
			Name:     "CatchallDenyWithAccount",
			Resource: svcResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "*",
					Resource: catchallResource,
					Access:   auth.AccessDenied,
				},
			},
			Error: auth.ErrForbidden,
		},
		{
			Name:     "CatchallDenyWithNoAccount",
			Resource: svcResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "*",
					Resource: catchallResource,
					Access:   auth.AccessDenied,
				},
			},
			Error: auth.ErrForbidden,
		},
		{
			Name:     "RulePriorityGrantFirst",
			Resource: svcResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "*",
					Resource: catchallResource,
					Access:   auth.AccessGranted,
					Priority: 1,
				},
				&auth.Rule{
					Scope:    "*",
					Resource: catchallResource,
					Access:   auth.AccessDenied,
					Priority: 0,
				},
			},
		},
		{
			Name:     "RulePriorityDenyFirst",
			Resource: svcResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "*",
					Resource: catchallResource,
					Access:   auth.AccessGranted,
					Priority: 0,
				},
				&auth.Rule{
					Scope:    "*",
					Resource: catchallResource,
					Access:   auth.AccessDenied,
					Priority: 1,
				},
			},
			Error: auth.ErrForbidden,
		},
		{
			Name:     "WebExactEndpointValid",
			Resource: webResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope:    "*",
					Resource: webResource,
				},
			},
		},
		{
			Name:     "WebExactEndpointInalid",
			Resource: webResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope: "*",
					Resource: &auth.Resource{
						Type:     webResource.Type,
						Name:     webResource.Name,
						Endpoint: "invalidendpoint",
					},
				},
			},
			Error: auth.ErrForbidden,
		},
		{
			Name:     "WebWildcardEndpoint",
			Resource: webResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope: "*",
					Resource: &auth.Resource{
						Type:     webResource.Type,
						Name:     webResource.Name,
						Endpoint: "*",
					},
				},
			},
		},
		{
			Name:     "WebWildcardPathEndpointValid",
			Resource: webResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope: "*",
					Resource: &auth.Resource{
						Type:     webResource.Type,
						Name:     webResource.Name,
						Endpoint: "/foo/*",
					},
				},
			},
		},
		{
			Name:     "WebWildcardPathEndpointInvalid",
			Resource: webResource,
			Account:  &auth.Account{},
			Rules: []*auth.Rule{
				&auth.Rule{
					Scope: "*",
					Resource: &auth.Resource{
						Type:     webResource.Type,
						Name:     webResource.Name,
						Endpoint: "/bar/*",
					},
				},
			},
			Error: auth.ErrForbidden,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			if err := Verify(tc.Rules, tc.Account, tc.Resource); err != tc.Error {
				t.Errorf("Expected %v but got %v", tc.Error, err)
			}
		})
	}
}
