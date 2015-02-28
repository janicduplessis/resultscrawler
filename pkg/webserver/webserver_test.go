package webserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/janicduplessis/resultscrawler/pkg/api"
	"github.com/janicduplessis/resultscrawler/pkg/store/fakestore"
)

const (
	testRSAPrivate = `-----BEGIN PRIVATE KEY-----
MIIEogIBAAKCAQEAwrxPf15a5MkYESPbRLbFXwsHqMjim6QF6LPLR5dmqEp6rqNW
A6CnMI1IAA0yGuy0ukPOvLiilouhcWJ0xnw++Gz7fraXp0P7C0GIvrdWTqFv+Qh3
+b69jGnSJwMKTeFUvBeNOjV3IzlyiucTgUn0kv1TPfwF5OksIVLVmaIaoRxxRMxQ
VrSQbz0UhaRJSveLDIJjKCpsEeM1pmTtxFctgJCuwSza4MLITj6LGWdls8w6nodC
D0QnM/p4VddTFsivYgwLjwzzk5XH6iMhYzOco+8ew3Y3p8GKHARB0K/jmPAdcZu/
wPdPyPWDOXvEKmelfI5q75Vj0EFmt2+P4RmTkwIDAQABAoIBABW1fH9Me4GJ0X8H
qkgMwBAKYL42Ntz2+hmpAX5nqHAWbXrOhqY84KaO+XnX/r/1p2gkawWq56U0x7im
KzJ9Y1+6dob3wAxLjc8BbUcllR+K67qtcQKMewEOQvlKY3mvJw0Y6wuULkXk/5nw
jMIbBoLkbsU4NUgBnoPQgjNwWNug6EY3KUGNgFZBoMrF5WPqfQGPuFsieTIo+gjb
2uj/UP8IvblpTMuPqUE7NVUwfQqLyRPhBb+KebE6YYynOBfZP7bTBGC7DcOVe6h3
QpNYp36sx5INJqO957B5lfPF6zjW+cCe/9ZiH4NaVEgPCtIVGY03HkP6DDCy2nJO
rkFf75ECgYEA5/SQpgMArnbUyQ0YVA3rTR07XTuINNNR7FWy0AAfbsjkkF13tyHz
FcCeWREOrclByxGVh0l0A7kCfVC7vL22UlnImRydrAkhhvTUuX0GUPfY2Rjss5il
qDBp3V1bc2++C7AHfjTlauwTzEZ29UlcZt0iXMyYpUHVE4riY0UlcP8CgYEA1uwK
EVGS4RFbXQv4OZKMhmUgIck26aTyjLzDFWWsN5eVzpBP8BERiyMosLjh/vC/tZJW
PIo828ytvT5n4KotOBkOkKj58wC0VTaEnF4KKAZ609P/ty+JXc6vipid2LFaWCFH
/lzWeGX2Zp/tpq8mcsAyHiV/yXZrX611QelQiW0CgYB50NTOerE27q1dUQU/z3eN
rhZpJkSoCXrytScNWaMoWVTABHZEtQ2mlNwURoMA/bsR3JA81nSZJ+aIzYdq3e8M
XJ6e2opruPfkmlvFdkWE7ETz7sUQpNAK/jH60Xafr0WNecrVmw4JEyZql28N7pMa
anQLbF+WGna+pqeyHrRFHQKBgBBFutVo2bgUulgnKdoiEGW0jmRAedni1UJ2oEak
dg+XeI41OvgwMqXYOaJ3vRSyYbF7rO/Uf5scuiLT8MV/3QCcVQ/620Hc0cqJ4Cx+
qkIxi2cya/AQt1PU7FGQEJNxiieWDX9ixBJFlgxbG4E9Tanuh1zk9fHo92Q9G92r
rp9ZAoGAeiX2huuFCeQ92nDc4CE9O1RWM/jHxQCs1MflTXMipm3O4Y2i4gQMTez2
kT6xp4rzXhjBRCuPoMXCGPnAFKd3NnjWbjwcESG/PgG1XG8M/GyZveWHKU0g3IRY
ElYGSP02sQk3nGiV5bxi8ikgXjoc1XsrWSUqvYfN2pkG9eXpgGs=
-----END PRIVATE KEY-----`
	testRSAPublic = `-----BEGIN PUBLIC KEY-----
AAAAB3NzaC1yc2EAAAADAQABAAABAQDCvE9/XlrkyRgRI9tEtsVfCweoyOKbpAXos8tHl2aoSnquo1YDoKcwjUgADTIa7LS6Q868uKKWi6FxYnTGfD74bPt+tpenQ/sLQYi+t1ZOoW/5CHf5vr2MadInAwpN4VS8F406NXcjOXKK5xOBSfSS/VM9/AXk6SwhUtWZohqhHHFEzFBWtJBvPRSFpElK94sMgmMoKmwR4zWmZO3EVy2AkK7BLNrgwshOPosZZ2WzzDqeh0IPRCcz+nhV11MWyK9iDAuPDPOTlcfqIyFjM5yj7x7DdjenwYocBEHQr+OY8B1xm7/A90/I9YM5e8QqZ6V8jmrvlWPQQWa3b4/hGZOT
-----END PUBLIC KEY-----`
)

func TestLoginInvalid(t *testing.T) {
	ts, _ := initServer()
	defer ts.Close()

	request := loginRequest{
		Email:      "lala@hotmail.com",
		Password:   "secret",
		DeviceType: 1,
	}
	res, err := post(ts.URL+urlLogin, request)
	if err != nil {
		t.Error(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Bad status code %d, should be %d", res.StatusCode, http.StatusOK)
		return
	}
	response := &loginResponse{}
	if err = parse(res, response); err != nil {
		t.Error(err)
		return
	}
	if response.Status != statusInvalidLogin || response.Token != "" || response.User != nil {
		t.Errorf("Unexpected response %+v, expected {Status:1 Token: User:<nil>}", response)
	}
}

func TestLoginGood(t *testing.T) {
	ts, webserver := initServer()
	defer ts.Close()

	webserver.userStore.CreateUser(&api.User{
		Email: "420blazeit@gmail.com",
	}, "h4ck3r")

	request := loginRequest{
		Email:      "420blazeit@gmail.com",
		Password:   "h4ck3r",
		DeviceType: 1,
	}
	res, err := post(ts.URL+urlLogin, request)
	if err != nil {
		t.Error(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Bad status code %d, should be %d", res.StatusCode, http.StatusOK)
		return
	}
	response := &loginResponse{}
	if err = parse(res, response); err != nil {
		t.Error(err)
		return
	}
	if response.Status != statusOK || response.Token == "" || response.User == nil {
		t.Errorf("Unexpected response %+v, expected {Status:0 Token:'Some value' User:<notnil>", response)
	}
}

type FakeCrawlerClient struct {
}

func (c *FakeCrawlerClient) Refresh(userID string) error {
	return nil
}

func initServer() (*httptest.Server, *Webserver) {
	store := new(fakestore.FakeStore)
	store.Data = make(map[string]*fakestore.TestUser)
	webserver := NewWebserver(&Config{
		UserStore:          store,
		CrawlerConfigStore: store,
		UserResultsStore:   store,
		RSAPublic:          []byte(testRSAPublic),
		RSAPrivate:         []byte(testRSAPrivate),
		CrawlerClient:      &FakeCrawlerClient{},
	})
	return httptest.NewServer(webserver.router), webserver
}

func post(url string, obj interface{}) (*http.Response, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return http.Post(url, "json", bytes.NewReader(data))
}

func parse(res *http.Response, obj interface{}) error {
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, obj); err != nil {
		return err
	}
	return nil
}
