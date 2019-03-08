package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/cblomart/registry-cleanup/responses/hub"
	"github.com/cblomart/registry-cleanup/rest"
)

const (
	//DefaultRegistry is the default value for the registry
	DefaultRegistry = "https://hub.docker.com"
	//DefaultRegex is the default match for tags to clean
	DefaultRegex = "^[A-Fa-f0-9]+$"
	//DefaultMin is the default value for the number of tags/images to keep
	DefaultMin = 3
	//DefaultMax is the maxium age of images in the registry
	DefaultMax = "15d"
	//HubPageSize docker hub page size
	HubPageSize = 100
)

type (
	//Repo repository informations
	Repo struct {
		FullName string
		Owner    string
	}

	//Author author informations
	Author struct {
		Name   string
		Email  string
		Avatar string
	}

	//Config plugin configuration informations
	Config struct {
		Username string
		Password string
		Repo     string
		Registry string
		Regex    string
		Min      int
		Max      time.Duration
	}

	//Plugin plugin data
	Plugin struct {
		Repo   Repo
		Config Config
	}
)

//Init set default values to config
func (c *Config) Init() error {
	if len(c.Registry) == 0 {
		c.Registry = DefaultRegistry
	}
	if len(c.Regex) == 0 {
		c.Regex = DefaultRegex
	}
	if c.Min == 0 {
		c.Min = DefaultMin
	}
	if c.Max.Seconds() == 0 {
		d, err := time.ParseDuration(DefaultMax)
		if err != nil {
			return fmt.Errorf("invalud default duration provided (%s)", DefaultMax)
		}
		c.Max = d
	}
	return nil
}

//Check the config values
func (c *Config) Check() error {
	// direct validation
	if len(c.Username) == 0 {
		return fmt.Errorf("empty username provided")
	}
	if len(c.Password) == 0 {
		return fmt.Errorf("empty password provided")
	}
	if len(c.Registry) == 0 {
		return fmt.Errorf("no registry provided")
	}
	if len(c.Repo) == 0 {
		return fmt.Errorf("no repository provided")
	}
	if len(c.Regex) == 0 {
		return fmt.Errorf("no regex match provided")
	}
	if c.Min == 0 {
		return fmt.Errorf("no minimum ammount of images/tags to keep")
	}
	if c.Max.Seconds() == 0 {
		return fmt.Errorf("no maximum age provided")
	}
	// complex validations
	// check registry
	_, err := url.Parse(c.Registry)
	if err != nil {
		return fmt.Errorf("registry is not in url format (%s)", c.Registry)
	}
	// check Regex
	_, err = regexp.Compile(c.Regex)
	if err != nil {
		return fmt.Errorf("invalid regex provided (%s)", c.Regex)
	}
	// avoid full match
	if c.Regex == ".*" || c.Regex == "^.*$" {
		return fmt.Errorf("regex would match everything (%s)", c.Regex)
	}
	return nil
}

//Exec executes the registry-cleanup plugin
func (p Plugin) Exec() error {
	// set default paramters
	err := p.Config.Init()
	if err != nil {
		return err
	}
	// check paramaters
	err = p.Config.Check()
	if err != nil {
		return err
	}
	if p.Config.Registry == DefaultRegistry {
		return p.ExecHub()
	}
	return p.ExecRegistry()
}

//ExecHub executes the registry-cleanup plugin on the docker hub
func (p Plugin) ExecHub() error {
	// get the base url
	baseurl := fmt.Sprintf("%s/v2/", p.Config.Registry)
	// initialize rest client
	r := rest.NewClient()
	// get a token
	data, err := r.Post(fmt.Sprintf("%s/users/login/", baseurl), map[string]string{"username": p.Config.Username, "password": p.Config.Password})
	if err != nil {
		return fmt.Errorf("could not get token")
	}
	var token hub.Token
	err = json.Unmarshal(data, &token)
	if err != nil {
		return fmt.Errorf("cannot read token from response")
	}
	r.Headers["Authorization"] = token.Token
	// get the tag list
	var tags []hub.Tag
	re := regexp.MustCompile(p.Config.Regex)
	var tagpage hub.Tags
	tagpage.Next = fmt.Sprintf("%s/repositories/%s/tag/?page_size=%d&page=%d", baseurl, p.Config.Repo, HubPageSize, 1)
	// loop trought the result pages
	for len(tagpage.Next) > 0 {
		data, err = r.Get(tagpage.Next)
		err = json.Unmarshal(data, &tagpage)
		if err != nil {
			return fmt.Errorf("cannot read tag page response")
		}
		for _, tag := range tagpage.Results {
			if !re.MatchString(tag.Name) {
				continue
			}
			tags = append(tags, tag)
		}
	}
	// order tags per date (newer to older)
	sort.SliceStable(tags, func(i, j int) bool {
		return tags[i].LastUpdated.After(tags[j].LastUpdated)
	})
	// parse the tags in reverse order to decice which to delete
	treshold := time.Now().Add(-p.Config.Max)
	var wg sync.WaitGroup
	deleting := false
	for i := len(tags) - 1; i >= 0; i-- {
		// stop if reached the minimum limit
		if i <= p.Config.Min-1 {
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
			go func() {
				_, err := r.Delete(fmt.Sprintf("%s/repositories/%s/tag/%s", baseurl, p.Config.Repo, tags[i].Name))
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s updated at '%s' error deleting", tags[i].Name, tags[i].LastUpdated.Format(time.RFC822))
				} else {
					fmt.Printf("%s updated at '%s' deleted", tags[i].Name, tags[i].LastUpdated.Format(time.RFC822))
				}
				wg.Done()
			}()
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
	return nil
}
