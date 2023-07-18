package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"gopkg.in/jcmturner/gokrb5.v7/client"
	"gopkg.in/jcmturner/gokrb5.v7/config"
	"gopkg.in/jcmturner/gokrb5.v7/keytab"
	"gopkg.in/jcmturner/gokrb5.v7/spnego"
)

func createKerberosClientWithPassword(pricipal string, password string) (*client.Client, error) {

	// Load the client krb5 config
	cfg, err := config.Load("/etc/krb5.conf")

	if err != nil {
		return nil, fmt.Errorf("failed to extract username and realm from pricipal")

	}

	username, realm := extractUsernameAndRealm(pricipal)

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

func createKerberosClientWithKeytab(ktPath string, pricipal string) (*client.Client, error) {
	// https://github.com/jcmturner/gokrb5/blob/855dbc707a37a21467aef6c0245fcf3328dc39ed/USAGE.md?plain=1#L20
	kt, err := keytab.Load(ktPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load keytab file: %v", err)
	}

	krb5Conf, err := config.Load("/etc/krb5.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to load Kerberos config: %v", err)
	}

	username, realm := extractUsernameAndRealm(pricipal)

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

func extractUsernameAndRealm(pricipal string) (string, string) {
	parts := strings.Split(pricipal, "@")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func makeAuthenticatedRequest(client *client.Client, url string, fqdn string) []byte {

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("could not create request: %v", err)
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
