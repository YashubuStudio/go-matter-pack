// Copyright (C) 2024 The go-matter Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package matter

import (
	"fmt"
	"net"
)

type query struct {
	payload          OnboardingPayload
	onNetworkAddress net.IP
	onNetworkPort    int
}

// QueryOption represents an option for creating a Query.
type QueryOption func(*query)

// WithQueryOnboardingPayload sets the onboarding payload for the query.
func WithQueryOnboardingPayload(payload OnboardingPayload) QueryOption {
	return func(q *query) {
		q.payload = payload
	}
}

// WithQueryOnNetworkAddress sets the on-network address for the query.
func WithQueryOnNetworkAddress(address net.IP, port int) QueryOption {
	return func(q *query) {
		q.onNetworkAddress = address
		q.onNetworkPort = port
	}
}

// NewQuery creates a new Query instance.
func NewQuery(opts ...QueryOption) Query {
	q := &query{
		payload: nil,
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

// OnboardingPayload returns the onboarding payload of the query.
func (q *query) OnboardingPayload() (OnboardingPayload, bool) {
	if q.payload == nil {
		return nil, false
	}
	return q.payload, true
}

// OnNetworkAddress returns the on-network address of the query.
func (q *query) OnNetworkAddress() (net.IP, int, bool) {
	if q.onNetworkAddress == nil {
		return nil, 0, false
	}
	return q.onNetworkAddress, q.onNetworkPort, true
}

// String returns the string representation of the query.
func (q *query) String() string {
	var base string
	if q.payload != nil {
		base = q.payload.String()
	}
	if q.onNetworkAddress == nil {
		return base
	}
	if base == "" {
		return net.JoinHostPort(q.onNetworkAddress.String(), fmt.Sprintf("%d", q.onNetworkPort))
	}
	return fmt.Sprintf("%s@%s:%d", base, q.onNetworkAddress.String(), q.onNetworkPort)
}
