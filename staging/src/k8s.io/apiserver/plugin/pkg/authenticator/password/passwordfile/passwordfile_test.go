/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package passwordfile

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/component-base/metrics/testutil"
)

func TestPasswordFile(t *testing.T) {
	auth, err := newWithContents(t, `
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",user1,uid1
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$+qPTPJDor7053pDbikasohiddOrYmPTmbLmJDF2d5u8",user2,uid2
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$L58wF8kR6VbBcZh9hZE4Sg+/wfa4RIN4I48IumrezeQ",user3,uid3,"group1,group2"
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$cy0VBOu3ySF7zhimgpnoTcfY0ZymxfV1dK3+KZCExz8",user4,uid4,"group2"
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KUpuLXcngRNmGGNxVU9YY9al/6ymX5N3rI0fQwD0C6M",user5,uid5,group5
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$OwkXaowr3+OGSBabJzEtzQom7DlMbT9/fuPfShycxXI",user6,uid6,group5,otherdata
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$VF+WtycA//hSxB8ygFVLRfJzUh47ISNwtK3qHDTn8lg",user7,uid7,"group1,group2",otherdata
`)
	if err != nil {
		t.Fatalf("unable to read passwordfile: %v", err)
	}

	testCases := []struct {
		desc     string
		username string
		password string
		wantResp *authenticator.Response
		ok       bool
		err      bool
	}{
		{
			desc:     "success user1",
			username: "user1",
			password: "password1",
			wantResp: &authenticator.Response{User: &user.DefaultInfo{Name: "user1", UID: "uid1"}},
			ok:       true,
		},
		{
			desc:     "success user2",
			username: "user2",
			password: "password2",
			wantResp: &authenticator.Response{User: &user.DefaultInfo{Name: "user2", UID: "uid2"}},
			ok:       true,
		},
		{
			desc:     "failure user1 incorrect password",
			username: "user1",
			password: "password2",
		},
		{
			desc:     "failure user2 incorrect password",
			username: "user2",
			password: "password1",
		},
		{
			desc:     "success user3",
			username: "user3",
			password: "password3",
			wantResp: &authenticator.Response{User: &user.DefaultInfo{Name: "user3", UID: "uid3", Groups: []string{"group1", "group2"}}},
			ok:       true,
		},
		{
			desc:     "success user4",
			username: "user4",
			password: "password4",
			wantResp: &authenticator.Response{User: &user.DefaultInfo{Name: "user4", UID: "uid4", Groups: []string{"group2"}}},
			ok:       true,
		},
		{
			desc:     "success user5",
			username: "user5",
			password: "password5",
			wantResp: &authenticator.Response{User: &user.DefaultInfo{Name: "user5", UID: "uid5", Groups: []string{"group5"}}},
			ok:       true,
		},
		{
			desc:     "success user6",
			username: "user6",
			password: "password6",
			wantResp: &authenticator.Response{User: &user.DefaultInfo{Name: "user6", UID: "uid6", Groups: []string{"group5"}}},
			ok:       true,
		},
		{
			desc:     "success user7",
			username: "user7",
			password: "password7",
			wantResp: &authenticator.Response{User: &user.DefaultInfo{Name: "user7", UID: "uid7", Groups: []string{"group1", "group2"}}},
			ok:       true,
		},
		{
			desc:     "failure user7 incorrect password",
			username: "user7",
			password: "passwordbad",
		},
		{
			desc:     "failure userbad does not exist",
			username: "userbad",
			password: "password7",
		},
		{
			desc:     "failure user8 does not exist",
			username: "user8",
			password: "password8",
		},
		{
			desc:     "failure as password is empty",
			username: "user1",
			password: "",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			resp, ok, err := auth.AuthenticatePassword(context.Background(), testCase.username, testCase.password)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if testCase.ok != ok {
				t.Errorf("got auth %v, want %v", ok, testCase.ok)
			}
			if diff := cmp.Diff(testCase.wantResp, resp); diff != "" {
				t.Errorf("got resp: %v, want resp: %v, diff: %s", resp, testCase.wantResp, diff)
			}
		})
	}
}

func TestBadPasswordFile(t *testing.T) {
	if _, err := newWithContents(t, `
password1,user1,uid1
password2,user2,uid2
password3,user3
password4
`); err == nil {
		t.Fatalf("unexpected non error")
	}
}

func TestInsufficientColumnsPasswordFile(t *testing.T) {
	if _, err := newWithContents(t, "password4\n"); err == nil {
		t.Fatalf("unexpected non error")
	}
}

func TestMetrics(t *testing.T) {
	testCases := []struct {
		desc         string
		fileContents string
		want         string
		user         string
		password     string
	}{
		{
			desc: "user not found",
			fileContents: `
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",user1,uid1
`,
			want: `
# HELP passwordfile_authentication_attempts_total [ALPHA] Count of requests that authenticate with basic authentication based on status.
# TYPE passwordfile_authentication_attempts_total counter
passwordfile_authentication_attempts_total{status="user_not_found"} 1
`,
			user:     "user2",
			password: "password1",
		},
		{
			desc: "success",
			fileContents: `
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",user1,uid1
`,
			want: `
# HELP passwordfile_authentication_attempts_total [ALPHA] Count of requests that authenticate with basic authentication based on status.
# TYPE passwordfile_authentication_attempts_total counter
passwordfile_authentication_attempts_total{status="success"} 1
`,
			user:     "user1",
			password: "password1",
		},
		{
			desc: "failure",
			fileContents: `
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",user1,uid1
`,
			want: `
# HELP passwordfile_authentication_attempts_total [ALPHA] Count of requests that authenticate with basic authentication based on status.
# TYPE passwordfile_authentication_attempts_total counter
passwordfile_authentication_attempts_total{status="failure"} 1
`,
			user:     "user1",
			password: "password2",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			auth, err := newWithContents(t, tc.fileContents)
			if err != nil {
				t.Fatal(err)
			}
			authenticatePasswordCounter.Reset()
			auth.AuthenticatePassword(context.Background(), tc.user, tc.password)
			if err := testutil.GatherAndCompare(legacyregistry.DefaultGatherer, strings.NewReader(tc.want), "passwordfile_authentication_attempts_total"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNewWithContents(t *testing.T) {
	base64PasswordHash := "KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ"
	base64Salt := "c2FsdHNhbHRzYWx0c2FsdA"
	passwordHash, err := base64.RawStdEncoding.DecodeString(base64PasswordHash)
	if err != nil {
		t.Fatalf("failed to decode password hash with error: %v", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(base64Salt)
	if err != nil {
		t.Fatalf("failed to decode salt with error: %v", err)
	}
	testCases := []struct {
		desc         string
		fileContents string
		want         *PasswordAuthenticator
	}{
		{
			desc: "hash only",
			fileContents: `
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",user1,uid1
`,
			want: &PasswordAuthenticator{
				users: map[string]*userPasswordInfo{
					"user1": {
						info: &user.DefaultInfo{
							Name: "user1",
							UID:  "uid1",
						},
						passwordChecker: &argon2IDChecker{
							passwordHash: passwordHash,
							salt:         salt,
							iterations:   1,
							memory:       65536,
							threads:      8,
						},
					},
				},
			},
		},
		{
			desc: "duplicate entries",
			fileContents: `
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",user1,uid1
"$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",user1,uid2
`,
			want: &PasswordAuthenticator{
				users: map[string]*userPasswordInfo{
					"user1": {
						info: &user.DefaultInfo{
							Name: "user1",
							UID:  "uid2",
						},
						passwordChecker: &argon2IDChecker{
							passwordHash: passwordHash,
							salt:         salt,
							iterations:   1,
							memory:       65536,
							threads:      8,
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			auth, err := newWithContents(t, tc.fileContents)
			if err != nil {
				t.Fatalf("newWithContents(%s) got error:%#v, want error: nil", tc.fileContents, err)
			}
			if diff := cmp.Diff(auth, tc.want, cmp.AllowUnexported(PasswordAuthenticator{}, userPasswordInfo{}, user.DefaultInfo{}, argon2IDChecker{})); diff != "" {
				t.Errorf("got PasswordAuthenticator:%#v, want: %#v, diff: %s", auth, tc.want, diff)
			}
		})
	}
}

func TestNewArgon2IDChecker(t *testing.T) {
	base64PasswordHash := "KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ"
	base64Salt := "c2FsdHNhbHRzYWx0c2FsdA"
	expectedPasswordHash, err := base64.RawStdEncoding.DecodeString(base64PasswordHash)
	if err != nil {
		t.Fatalf("failed to decode password hash with error: %v", err)
	}
	expectedSalt, err := base64.RawStdEncoding.DecodeString(base64Salt)
	if err != nil {
		t.Fatalf("failed to decode salt with error: %v", err)
	}

	testCases := []struct {
		desc            string
		encoded         string
		passwordChecker passwordChecker
		wantError       bool
	}{
		{
			desc:    "success",
			encoded: "$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			passwordChecker: &argon2IDChecker{
				passwordHash: expectedPasswordHash,
				salt:         expectedSalt,
				iterations:   1,
				memory:       65536,
				threads:      8,
			},
		},
		{
			desc:      "details does not start with $",
			encoded:   "argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "version not a number",
			encoded:   "$argon2id$v=a$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "memory not a number",
			encoded:   "$argon2id$v=19$m=a,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "iterations not a number",
			encoded:   "$argon2id$v=19$m=65536,t=a,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "threads not a number",
			encoded:   "$argon2id$v=19$m=65536,t=1,p=a$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "deatails contains extra params",
			encoded:   "$argon2id$v=19$m=65536,t=1,p=8,q=1212$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "deatails contains less parameters",
			encoded:   "$argon2id$v=19$m=65536,t=1,q=1212$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "not the correct version",
			encoded:   "$argon2id$v=20$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "memory not in uint32 range",
			encoded:   "$argon2id$v=19$m=4294967296,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "iterations not in uint32 range",
			encoded:   "$argon2id$v=19$m=65536,t=4294967296,p=8$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "threads not in uint8 range",
			encoded:   "$argon2id$v=19$m=65536,t=1,p=256$c2FsdHNhbHRzYWx0c2FsdA$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "hash length is below 32 bytes",
			encoded:   "$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$cGFzcw",
			wantError: true,
		},
		{
			desc:      "salt length is below 16 bytes",
			encoded:   "$argon2id$v=19$m=65536,t=1,p=8$cGFzcw$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
		{
			desc:      "hash invalid base64 encoding",
			encoded:   "$argon2id$v=19$m=65536,t=1,p=8$c2FsdHNhbHRzYWx0c2FsdA$MFQWCYLBMFQWCYLBMFQWCYLBMFQWCYLBMFQWCYLBMFQWCYLBMFQWC",
			wantError: true,
		},
		{
			desc:      "salt invalid base64 encoding",
			encoded:   "$argon2id$v=19$m=65536,t=1,p=8$MFQWCYLBMFQWCYLBMFQWCYLBMFQWCYLBMFQWCYLBMFQWCYLBMFQWC$KQg2lkQo1KzHByE8ZajR+UuHaaX5GnwF7OJ0mZfYlMQ",
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := newArgon2IDChecker(tc.encoded)
			if tc.wantError {
				if err == nil {
					t.Errorf("newArgon2IDChecker(%q), got err: nil, want err: non-nil", tc.encoded)
				}
			} else if err != nil {
				t.Errorf("newArgon2IDChecker(%q), got err: %v, want err: nil", tc.encoded, err)
			} else if diff := cmp.Diff(got, tc.passwordChecker, cmp.AllowUnexported(argon2IDChecker{})); diff != "" {
				t.Errorf("newArgon2IDChecker(%q) got passwordChecker: %#v, want: %#v, diff:%s", tc.encoded, got, tc.passwordChecker, diff)
			}
		})
	}
}

func newWithContents(t *testing.T, contents string) (auth *PasswordAuthenticator, err error) {
	f, err := ioutil.TempFile("", "passwordfile_test")
	if err != nil {
		t.Fatalf("unexpected error creating passwordfile: %v", err)
	}
	f.Close()
	defer os.Remove(f.Name())

	if err := ioutil.WriteFile(f.Name(), []byte(contents), 0700); err != nil {
		t.Fatalf("unexpected error writing passwordfile: %v", err)
	}

	return NewCSV(f.Name())
}
