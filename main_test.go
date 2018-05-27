package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_UrlParseQuery(t *testing.T) {
	q := `token=gIkuvaNzQIHg97ATvDxqgjtO&team_id=T0001&team_domain=example&enterprise_id=E0001&enterprise_name=Globular%2520Construct%2520Inc&channel_id=C2147483705&channel_name=test&user_id=U2147483697&user_name=Steve&command=%2Fweather&text=94070&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2F1234%2F5678&trigger_id=13345224609.738474920.8088930838d88f008e0`

	v, err := url.ParseQuery(q)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, "gIkuvaNzQIHg97ATvDxqgjtO", v.Get("token"))
}
