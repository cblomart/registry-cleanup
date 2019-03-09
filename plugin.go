package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cblomart/registry-cleanup/responses/registry"

	"github.com/cblomart/registry-cleanup/responses/hub"
	"github.com/cblomart/registry-cleanup/rest"
)

const (
	//DefaultRegistry is the default value for the registry
	DefaultRegistry = "https://cloud.docker.com"
	//HubPageSize docker hub page size
	HubPageSize = 100
	//RegistryAuthHeader registry authentication header
	RegistryAuthHeader = "Www-Authenticate"
	//Regex to  validate auth header
	ValidAuthHeader = "^[Bb]earer *((realm|service|scope|error)=\"[A-Za-z0-9-_./:]+\",?){2,4}$"
	//Required Registry Scope to delete tags
	RegistryScope = "pull,push"
)

type (
	//Plugin plugin data
	Plugin struct {
		Username string
		Password string
		Repo     string
		Registry string
		Regex    string
		Min      int
		Max      time.Duration
		Verbose  bool
		DryRun   bool
	}
)

//Check the config values
func (p *Plugin) Check() error {
	// direct validation
	if len(p.Username) == 0 {
		return fmt.Errorf("empty username provided")
	}
	if len(p.Password) == 0 {
		return fmt.Errorf("empty password provided")
	}
	if len(p.Registry) == 0 {
		return fmt.Errorf("no registry provided")
	}
	if len(p.Repo) == 0 {
		return fmt.Errorf("no repository provided")
	}
	if len(p.Regex) == 0 {
		return fmt.Errorf("no regex match provided")
	}
	if p.Min == 0 {
		return fmt.Errorf("no minimum ammount of images/tags to keep")
	}
	if p.Max.Seconds() == 0 {
		return fmt.Errorf("no maximum age provided")
	}
	// complex validations
	// check registry
	_, err := url.Parse(p.Registry)
	if err != nil {
		return fmt.Errorf("registry is not in url format (%s)", p.Registry)
	}
	// check Regex
	_, err = regexp.Compile(p.Regex)
	if err != nil {
		return fmt.Errorf("invalid regex provided (%s)", p.Regex)
	}
	// avoid full match
	if p.Regex == ".*" || p.Regex == "^.*$" {
		return fmt.Errorf("regex would match everything (%s)", p.Regex)
	}
	return nil
}

//Exec executes the registry-cleanup plugin
func (p Plugin) Exec() error {
	// check paramaters
	err := p.Check()
	if err != nil {
		return err
	}
	// if default registry use docker hub api
	if p.Registry == DefaultRegistry {
		return p.ExecHub()
	}
	// else use registry api
	return p.ExecRegistry()
}

//ExecHub executes the registry-cleanup plugin on the docker hub
func (p Plugin) ExecHub() error {
	// get the base url
	baseurl := fmt.Sprintf("%s/v2/", p.Registry)
	// initialize rest client
	r := rest.NewClient()
	// get a token
	data, err := r.Post(fmt.Sprintf("%susers/login/", baseurl), map[string]string{"username": p.Username, "password": p.Password})
	if err != nil {
		if p.Verbose {
			fmt.Println(err)
		}
		return fmt.Errorf("could not get token")
	}
	var token hub.Token
	err = json.Unmarshal(data, &token)
	if err != nil {
		if p.Verbose {
			fmt.Println(err)
		}
		return fmt.Errorf("cannot read token from response")
	}
	if p.Verbose {
		fmt.Printf("authenticated with %s\n", p.Username)
	}
	r.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token.Token)
	// get the tag list
	var tags []hub.Tag
	re := regexp.MustCompile(p.Regex)
	var tagpage hub.Tags
	tagpage.Next = fmt.Sprintf("%srepositories/%s/tags/?page_size=%d&page=%d", baseurl, p.Repo, HubPageSize, 1)
	// loop trought the result pages
	for len(tagpage.Next) > 0 {
		data, err = r.Get(tagpage.Next)
		if err != nil {
			if p.Verbose {
				fmt.Println(err)
			}
			return fmt.Errorf("cannot get tag page")
		}
		tagpage = hub.Tags{}
		err = json.Unmarshal(data, &tagpage)
		if err != nil {
			if p.Verbose {
				fmt.Println(err)
			}
			return fmt.Errorf("cannot read tag page response")
		}
		for _, tag := range tagpage.Results {
			if !re.MatchString(tag.Name) || tag.Name == "latest" {
				continue
			}
			tags = append(tags, tag)
		}
		if p.Verbose {
			fmt.Printf("found %d tags/images\n", len(tags))
		}
	}
	// order tags per date (newer to older)
	sort.SliceStable(tags, func(i, j int) bool {
		return tags[i].LastUpdated.After(tags[j].LastUpdated)
	})
	// parse the tags in reverse order to decice which to delete
	treshold := time.Now().Add(-p.Max)
	var wg sync.WaitGroup
	deleting := false
	for i := len(tags) - 1; i >= 0; i-- {
		// stop if reached the minimum limit
		if i <= p.Min-1 {
			break
		}
		// delete if older than treshold
		if tags[i].LastUpdated.Before(treshold) {
			wg.Add(1)
			// update the deleting flag
			if !deleting {
				deleting = true
			}
			// send the delete request async
			go func(tag hub.Tag) {
				defer wg.Done()
				if p.DryRun {
					fmt.Printf("%s:%s (%s) would be deleted\n", p.Repo, tag.Name, tag.LastUpdated.Format(time.RFC822))
					return
				}
				_, err := r.Delete(fmt.Sprintf("%srepositories/%s/tags/%s/", baseurl, p.Repo, tag.Name))
				if err != nil {
					if p.Verbose {
						fmt.Println(err)
					}
					fmt.Fprintf(os.Stderr, "%s:%s (%s) error deleting\n", p.Repo, tag.Name, tag.LastUpdated.Format(time.RFC822))
				}
				fmt.Printf("%s:%s (%s) deleted\n", p.Repo, tag.Name, tag.LastUpdated.Format(time.RFC822))
			}(tags[i])
		}
	}
	// if deleting woit for the results
	if deleting {
		wg.Wait()
	}
	return nil
}

//ExecRegistry executes the registry-cleanup plugin on a private registry
func (p Plugin) ExecRegistry() error {
	// initialize rest client
	r := rest.NewClient()
	// check v2
	rawheaders, err := r.Head(fmt.Sprintf("%s/v2/", p.Registry))
	if err != nil {
		return fmt.Errorf("%s does not support registry v2", p.Registry)
	}
	// check if headers recieved
	if len(rawheaders) == 0 {
		return fmt.Errorf("no header revieved from registry v2 endpoint")
	}
	// check the authentication endpoint
	var headers map[string][]string
	err = json.Unmarshal(rawheaders, &headers)
	if err != nil {
		if p.Verbose {
			fmt.Println(err)
		}
		return fmt.Errorf("could not deserialize headers")
	}
	// registry auth realm
	realm := ""
	// registry auth service
	service := ""
	if authheader, ok := headers[RegistryAuthHeader]; ok {
		if len(authheader) != 1 {
			return fmt.Errorf("more than one authentication header sent")
		}
		realm, service, _, err = decodeauthheader(authheader[0])
		if err != nil {
			return err
		}
	}
	// authenticate for registry
	userpass := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", p.Username, p.Password)))
	r.Headers["Authorization"] = fmt.Sprintf("Basic %s", userpass)
	data, err := r.Get(fmt.Sprintf("%s?service=%s&scope=%s:%s", realm, service, p.Repo, RegistryScope))
	if err != nil {
		if p.Verbose {
			fmt.Println(err)
		}
		return fmt.Errorf("could not get token")
	}
	var token registry.TokenResp
	err = json.Unmarshal(data, &token)
	if err != nil {
		if p.Verbose {
			fmt.Println(err)
		}
		return fmt.Errorf("could decode token")
	}
	// set authentication
	if p.Verbose {
		fmt.Printf("authenticated with %s\n", p.Username)
	}
	r.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token.Token)
	// get the manifest list

	return nil
}

// decode registry auth header
func decodeauthheader(header string) (string, string, string, error) {
	// registry auth realm
	realm := ""
	// registry auth service
	service := ""
	// registry required scope to delete tags
	scope := "push,pull"
	matched, err := regexp.MatchString(ValidAuthHeader, header)
	if err != nil {
		return realm, service, scope, fmt.Errorf("error validating auth header")
	}
	if !matched {
		return realm, service, scope, fmt.Errorf("invalid auth header")
	}
	parts := strings.Split(header, " ")
	rawfields := parts[len(parts)-1]
	fields := strings.Split(rawfields, ",")
	for _, field := range fields {
		elements := strings.Split(field, "=")
		switch elements[0] {
		case "realm":
			realm = elements[1][1 : len(elements[1])-1]
		case "service":
			service = elements[1][1 : len(elements[1])-1]
		case "scope":
			scope = elements[1][1 : len(elements[1])-1]
		}
	}
	return realm, service, scope, nil
}
