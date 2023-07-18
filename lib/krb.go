package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/jcmturner/gokrb5.v7/client"
	"gopkg.in/jcmturner/gokrb5.v7/config"
	"gopkg.in/jcmturner/gokrb5.v7/keytab"
	"gopkg.in/jcmturner/gokrb5.v7/spnego"
)

func CreateKerberosClientWithPassword(pricipal string, password string) (*client.Client, error) {

	// Load the client krb5 config
	cfg, err := config.Load("/etc/krb5.conf")

	if err != nil {
		return nil, fmt.Errorf("failed to extract username and realm from pricipal")

	}

	username, realm := ExtractUsernameAndRealm(pricipal)

	if username == "" {
		return nil, fmt.Errorf("failed to extract username and realm from pricipal")

	}

	cli := client.NewClientWithPassword(username, realm, password, cfg)

	// Log in the client
	err = cli.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to extract username and realm from pricipal")

	}

	return cli, nil
}

func CreateKerberosClientWithKeytab(ktPath string, pricipal string) (*client.Client, error) {
	// https://github.com/jcmturner/gokrb5/blob/855dbc707a37a21467aef6c0245fcf3328dc39ed/USAGE.md?plain=1#L20
	kt, err := keytab.Load(ktPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load keytab file: %v", err)
	}

	krb5Conf, err := config.Load("/etc/krb5.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to load Kerberos config: %v", err)
	}

	username, realm := ExtractUsernameAndRealm(pricipal)

	if username == "" {
		return nil, fmt.Errorf("failed to extract username and realm from pricipal")
	}

	cli := client.NewClientWithKeytab(username, realm, kt, krb5Conf)

	// Log in the client
	err = cli.Login()
	if err != nil {
		return nil, fmt.Errorf("failed to extract username and realm from pricipal")

	}

	return cli, nil
}

func ExtractUsernameAndRealm(pricipal string) (string, string) {
	parts := strings.Split(pricipal, "@")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func extractDomainFromURL(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	host := parsedURL.Hostname()

	return host, nil
}

func MakeKrb5Request(client *client.Client, url string) []byte {

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("could not create request: %v", err)
	}

	fqdn, err := extractDomainFromURL(url)
	if err != nil {
		log.Fatalf("could not extract fqdn from url: %v", err)
	}

	spn := fmt.Sprintf("HTTP/%s", fqdn)

	spnegoCl := spnego.NewClient(client, nil, spn)

	// Make the request
	resp, err := spnegoCl.Do(r)
	if err != nil {
		log.Fatalf("error making request: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %v", err)
	}
	// fmt.Println(string(body))

	defer resp.Body.Close()

	return body

}

func MakeKrb5RequestWithKeytab(ktPath string, pricipal string, url string) []byte {

	krb5cli, err := CreateKerberosClientWithKeytab(ktPath, pricipal)

	if err != nil {
		log.Fatalf("could not create krb5 client: %v", err)
	}

	return MakeKrb5Request(krb5cli, url)

}

func MakeKrb5RequestWithPassword(pricipal string, password string, url string) []byte {

	krb5cli, err := CreateKerberosClientWithPassword(pricipal, password)

	if err != nil {
		log.Fatalf("could not create krb5 client: %v", err)
	}

	return MakeKrb5Request(krb5cli, url)

}
