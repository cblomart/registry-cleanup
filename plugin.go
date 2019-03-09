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
		Dump     bool
	}

	//Tag tag data
	Tag struct {
		Name    string
		Created time.Time
		Digest  string
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
	r := rest.NewClient(p.Dump)
	// get a token
	var token hub.Token
	err := r.Post(fmt.Sprintf("%susers/login/", baseurl), map[string]string{"username": p.Username, "password": p.Password}, &token)
	if err != nil {
		if p.Verbose {
			fmt.Println(err)
		}
		return fmt.Errorf("could not get token")
	}
	if p.Verbose {
		fmt.Printf("authenticated with %s\n", p.Username)
	}
	r.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token.Token)
	// get the tag list
	var tags []hub.Tag
	re := regexp.MustCompile(p.Regex)
	url := fmt.Sprintf("%srepositories/%s/tags/?page_size=%d&page=%d", baseurl, p.Repo, HubPageSize, 1)
	var tagpage hub.Tags
	// loop trought the result pages
	for len(url) > 0 {
		tagpage = hub.Tags{}
		err = r.Get(url, nil, &tagpage)
		if err != nil {
			if p.Verbose {
				fmt.Println(err)
			}
			return fmt.Errorf("cannot get tag page")
		}
		url = tagpage.Next
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
	deleted := 0
	errors := 0
	for i := len(tags) - 1; i >= 0; i-- {
		// stop if reached the minimum limit
		if i <= p.Min-1 {
			break
		}
		// delete if older than treshold
		if tags[i].LastUpdated.Before(treshold) {
			wg.Add(1)
			// update the deleting flag
			// send the delete request async
			go func(tag hub.Tag) {
				defer wg.Done()
				if p.DryRun {
					fmt.Printf("dryrun [%s] %s:%s\n", tag.LastUpdated.Format(time.RFC822), p.Repo, tag.Name)
					errors++
					return
				}
				err := r.Delete(fmt.Sprintf("%srepositories/%s/tags/%s/", baseurl, p.Repo, tag.Name), nil, nil)
				if err != nil {
					if p.Verbose {
						fmt.Println(err)
					}
					fmt.Fprintf(os.Stderr, "error [%s] %s:%s\n", tag.LastUpdated.Format(time.RFC822), p.Repo, tag.Name)
					errors++
					return
				}
				deleted++
				fmt.Printf("deleted [%s] %s:%s\n", tag.LastUpdated.Format(time.RFC822), p.Repo, tag.Name)
			}(tags[i])
		}
	}
	// wait for the results
	wg.Wait()
	if errors > 0 {
		fmt.Printf("issue deleting %d tags/images\n", errors)
	}
	fmt.Printf("successfully deleted %d tags/images\n", deleted)
	return nil
}

//ExecRegistry executes the registry-cleanup plugin on a private registry
func (p Plugin) ExecRegistry() error {
	// set the base url
	baseurl := fmt.Sprintf("%s/v2/", p.Registry)
	// initialize rest client
	r := rest.NewClient(p.Dump)
	// check v2
	var headers map[string][]string
	err := r.Head(baseurl, nil, &headers)
	if err != nil {
		return fmt.Errorf("%s does not support registry v2", p.Registry)
	}
	// registry auth realm
	realm := ""
	// registry auth service
	service := ""
	if authheader, ok := headers[registry.AuthHeader]; ok {
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
	var token registry.TokenResp
	err = r.Get(fmt.Sprintf("%s?service=%s&scope=repository:%s:%s", realm, service, p.Repo, registry.Scope), nil, &token)
	if err != nil {
		if p.Verbose {
			fmt.Println(err)
		}
		return fmt.Errorf("could not get token")
	}
	// set authentication
	if p.Verbose {
		fmt.Printf("authenticated with %s\n", p.Username)
	}
	r.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token.Token)
	// get the tags list
	var tags registry.TagsListResp
	err = r.Get(fmt.Sprintf("%s%s/tags/list", baseurl, p.Repo), nil, &tags)
	if err != nil {
		if p.Verbose {
			fmt.Println(err)
		}
		return fmt.Errorf("could not get tag list")
	}
	// filter tags list
	var scopedTags []string
	re := regexp.MustCompile(p.Regex)
	for _, tag := range tags.Tags {
		if !re.MatchString(tag) || tag == "latest" {
			continue
		}
		scopedTags = append(scopedTags, tag)

	}
	if p.Verbose {
		fmt.Printf("found %d tags/images\n", len(scopedTags))
	}
	// set mime type for manifests
	r.Headers["Accept"] = registry.ManifestMimeV2
	// get informations on scoped tags
	var tagInfos []Tag
	var wg sync.WaitGroup
	wg.Add(len(scopedTags))
	for _, tag := range scopedTags {
		go func(tag string) {
			// defer completion
			defer wg.Done()
			// check version of the manifest
			var headers map[string][]string
			err := r.Head(fmt.Sprintf("%s%s/manifests/%s", baseurl, p.Repo, tag), nil, &headers)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not head manifest: %s\n", err)
				return
			}
			// get the digest from headers
			digest := ""
			if digests, ok := headers[registry.DigestHeader]; ok {
				digest = digests[0]
			}
			if len(digest) == 0 {
				fmt.Fprintf(os.Stderr, "no digest for manifest: %s\n", tag)
				return
			}
			// check manifest in function of version
			if mimetype, ok := headers["Content-Type"]; ok {
				switch mimetype[0] {
				case registry.ManifestMimeV2:
					var manifest registry.ManifestRespV2
					err = r.Get(fmt.Sprintf("%s%s/manifests/%s", baseurl, p.Repo, tag), nil, &manifest)
					if err != nil {
						fmt.Fprintf(os.Stderr, "could not get manifest: %s\n", err)
						return
					}
					var image registry.Image
					err = r.Get(fmt.Sprintf("%s%s/blobs/%s", baseurl, p.Repo, manifest.Config.Digest), nil, &image)
					if err != nil {
						fmt.Fprintf(os.Stderr, "could not get config blob: %s\n", err)
						return
					}
					tagInfos = append(tagInfos, Tag{Name: tag, Created: image.Created, Digest: digest})
				case registry.ManifestMimeV1:
					// get the manifest
					var manifest registry.ManifestRespV1
					err = r.Get(fmt.Sprintf("%s%s/manifests/%s", baseurl, p.Repo, tag), nil, &manifest)
					if err != nil {
						fmt.Fprintf(os.Stderr, "could not get manifest: %s\n", err)
						return
					}
					// get all images informations and check for the latest
					images := make([]registry.Image, len(manifest.History))
					latest := -1
					for i, h := range manifest.History {
						err = json.Unmarshal([]byte(h.V1Compatibility), &images[i])
						if err != nil {
							fmt.Fprintf(os.Stderr, "could not decode image from history: %s\n", err)
							continue
						}
						if latest == -1 {
							latest = i
							continue
						}
						if images[i].Created.After(images[latest].Created) {
							latest = i
						}
					}
					tagInfos = append(tagInfos, Tag{Name: tag, Created: images[latest].Created, Digest: digest})
				default:
					fmt.Printf("manifest type not handled for %s: %s\n", tag, mimetype[0])
				}
			}
		}(tag)
	}
	wg.Wait()
	// indicate the details found
	if p.Verbose {
		fmt.Printf("found details on %d tags/images\n", len(tagInfos))
	}
	// order tags infos per date (newer to older)
	sort.SliceStable(tagInfos, func(i, j int) bool {
		return tagInfos[i].Created.After(tagInfos[j].Created)
	})
	// parse the tags in reverse order to decice which to delete
	treshold := time.Now().Add(-p.Max)
	errors := 0
	deleted := 0
	for i := len(tagInfos) - 1; i >= 0; i-- {
		// stop if reached the minimum limit
		if i <= p.Min-1 {
			break
		}
		// delete if older than treshold
		if tagInfos[i].Created.Before(treshold) {
			wg.Add(1)
			// send the delete request async
			go func(tag Tag) {
				defer wg.Done()
				if p.DryRun {
					fmt.Printf("dryrun [%s] %s:%s\n", tag.Created.Format(time.RFC822), p.Repo, tag.Name)
					errors++
					return
				}
				err := r.Delete(fmt.Sprintf("%s%s/manifests/%s", baseurl, p.Repo, tag.Digest), nil, nil)
				if err != nil {
					if p.Verbose {
						fmt.Println(err)
					}
					fmt.Fprintf(os.Stderr, "error [%s] %s:%s\n", tag.Created.Format(time.RFC822), p.Repo, tag.Name)
					errors++
					return
				}
				deleted++
				fmt.Printf("deleted [%s] %s:%s\n", tag.Created.Format(time.RFC822), p.Repo, tag.Name)
			}(tagInfos[i])
		}
	}
	// if deleting wait for the results
	wg.Wait()
	if errors > 0 {
		fmt.Printf("issue deleting %d tags/images\n", errors)
	}
	fmt.Printf("successfully deleted %d tags/images\n", deleted)
	return nil
}

// decode registry auth header
func decodeauthheader(header string) (string, string, string, error) {
	// registry auth realm
	realm := ""
	// registry auth service
	service := ""
	// registry required scope to delete tags
	scope := registry.Scope
	matched, err := regexp.MatchString(registry.ValidAuthHeader, header)
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
